package eth

import "errors"

var (
	ErrInsufficientBalance = errors.New("account has insufficient funds to proceed")
)
