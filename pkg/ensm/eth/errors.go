package eth

import "errors"

var (
	ErrInsufficientBalance = errors.New("account has insufficient funds to proceed. fund the account and try again with the mnemonic")
)
