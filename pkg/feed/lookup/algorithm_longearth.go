package lookup

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type stepFunc func(ctx context.Context, t uint64, hint Epoch) interface{}

// LongEarthLookaheadDelay is the headstart the lookahead gives R before it launches
var LongEarthLookaheadDelay = 500 * time.Millisecond

// LongEarthLookbackDelay is the headstart the lookback gives R before it launches
var LongEarthLookbackDelay = 500 * time.Millisecond

// LongEarthAlgorithm explores possible lookup paths in parallel, pruning paths as soon
// as a more promising lookup path is found. As a result, this lookup algorithm is an order
// of magnitude faster than the FluzCapacitor algorithm, but at the expense of more exploratory reads.
// This algorithm works as follows. On each step, the next epoch is immediately looked up (R)
// and given a head start, while two parallel "steps" are launched a short time after:
// look ahead (A) is the path the algorithm would take if the R lookup returns a value, whereas
// look back (B) is the path the algorithm would take if the R lookup failed.
// as soon as R is actually finished, the A or B paths are pruned depending on the value of R.
// if A returns earlier than R, then R and B read operations can be safely canceled, saving time.
// The maximum number of active read operations is calculated as 2^(timeout/headstart).
// If headstart is infinite, this algorithm behaves as FluzCapacitor.
// timeout is the maximum execution time of the passed `read` function.
// the two head starts can be configured by changing LongEarthLookaheadDelay or LongEarthLookbackDelay
func LongEarthAlgorithm(ctx context.Context, now uint64, hint Epoch, read ReadFunc) (interface{}, error) {
	if hint == NoClue {
		hint = worstHint
	}

	var stepCounter int32 // for debugging, stepCounter allows to give an ID to each step instance

	errc := make(chan struct{}) // errc will help as an error shortcut signal
	var gerr error              // in case of error, this variable will be set

	var step stepFunc // For efficiency, the algorithm step is defined as a closure
	step = func(ctxS context.Context, t uint64, last Epoch) interface{} {
		stepID := atomic.AddInt32(&stepCounter, 1) // give an ID to this call instance
		trace(stepID, "init: t=%d, last=%s", t, last.String())
		var valueA, valueB, valueR interface{}
		var valueAMu sync.RWMutex
		var valueBMu sync.RWMutex
		var valueRMu sync.RWMutex

		// initialize the three read contexts
		ctxR, cancelR := context.WithCancel(ctxS) // will handle the current read operation
		ctxA, cancelA := context.WithCancel(ctxS) // will handle the lookahead path
		ctxB, cancelB := context.WithCancel(ctxS) // will handle the lookback path

		epoch := GetNextEpoch(last, t) // calculate the epoch to look up in this step instance

		// define the lookAhead function, which will follow the path as if R was successful
		lookAhead := func() {
			valuea := step(ctxA, t, epoch) // launch the next step, recursively.
			valueAMu.Lock()
			valueA = valuea
			valueAMu.Unlock()
			if valuea != nil { // if this path is successful, we don't need R or B.
				cancelB()
				cancelR()
			}

		}

		// define the lookBack function, which will follow the path as if R was unsuccessful
		lookBack := func() {
			if epoch.Base() == last.Base() {
				return
			}
			base := epoch.Base()
			if base == 0 { // skipcq: TCV-001
				return
			}
			valueb := step(ctxB, base-1, last)
			valueBMu.Lock()
			valueB = valueb
			valueBMu.Unlock()
		}

		go func() { //goroutine to read the current epoch (R)
			defer cancelR()
			var err error
			valuer, err := read(ctxR, epoch, now) // read this epoch
			valueRMu.Lock()
			valueR = valuer
			valueRMu.Unlock()
			if valuer == nil { // if unsuccessful, cancel lookahead, otherwise cancel lookback.
				cancelA()
			} else {
				cancelB()
				//cancelA() // cancel this also for faster eject
			}
			if err != nil && !errors.Is(err, context.Canceled) { // skipcq: TCV-001
				gerr = err
				close(errc)
			}
		}()

		go func() { // goroutine to give a headstart to R and then launch lookahead.
			defer cancelA()

			// if we are at the lowest level or the epoch to look up equals the last one,
			// then we cannot lookahead (can't go lower or repeat the same lookup, this would
			// cause an infinite loop)
			if epoch.Level == LowestLevel || epoch.Equals(last) {
				return
			}
			// give a head start to R, or launch immediately if R finishes early enough
			select {
			case <-TimeAfter(LongEarthLookaheadDelay): // skipcq: TCV-001
				lookAhead()
			case <-ctxR.Done():
				valueRMu.Lock()
				valuer := valueR
				valueRMu.Unlock()
				if valuer != nil {
					lookAhead() // only look ahead if R was successful
				}
			case <-ctxA.Done():
			}
		}()

		go func() { // goroutine to give a headstart to R and then launch lookback.
			defer cancelB()
			// give a head start to R, or launch immediately if R finishes early enough
			select {
			case <-TimeAfter(LongEarthLookbackDelay): // skipcq: TCV-001
				lookBack()
			case <-ctxR.Done():
				valueRMu.Lock()
				valuer := valueR
				valueRMu.Unlock()
				if valuer == nil {
					lookBack() // only look back in case R failed
				}
			case <-ctxB.Done():
			}
		}()

		<-ctxA.Done()
		valueAMu.Lock()
		valuea := valueA
		valueAMu.Unlock()
		if valuea != nil {
			trace(stepID, "Returning valueA=%v", valuea)
			return valuea
		}

		<-ctxR.Done()
		valueRMu.Lock()
		valuer := valueR
		valueRMu.Unlock()
		if valuer != nil {
			trace(stepID, "Returning valueR=%v", valuer)
			return valuer
		}

		<-ctxB.Done()
		valueBMu.Lock()
		valueb := valueB
		valueBMu.Unlock()
		trace(stepID, "Returning valueB=%v", valueb)
		return valueb
	}

	var value interface{}
	stepCtx, cancel := context.WithCancel(ctx)

	go func() { // launch the root step in its own goroutine to allow cancellation
		defer cancel()
		value = step(stepCtx, now, hint)
	}()

	// wait for the algorithm to finish, but shortcut in case
	// of errors
	select {
	case <-stepCtx.Done():
	case <-errc: // skipcq: TCV-001
		cancel()
		return nil, gerr
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if value != nil || hint == worstHint {
		return value, nil
	}

	// at this point the algorithm did not return a value,
	// so we challenge the hint given.
	value, err := read(ctx, hint, now)
	if err != nil {
		return nil, err
	}
	if value != nil {
		return value, nil // hint is valid, return it.
	}

	// hint is invalid. Invoke the algorithm
	// without hint.
	now = hint.Base()
	if hint.Level == HighestLevel { // skipcq: TCV-001
		now--
	}

	return LongEarthAlgorithm(ctx, now, NoClue, read)
}
