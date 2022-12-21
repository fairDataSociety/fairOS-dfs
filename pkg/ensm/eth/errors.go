package eth

import "errors"

var (
	// ErrInsufficientBalance is used to denote user account has low funds for ens registration
	ErrInsufficientBalance = errors.New("insufficient funds")
)
