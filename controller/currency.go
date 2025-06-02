package controller

import "errors"

type Currency string

const (
	CurrencyMonero Currency = "xmr"
)

var ErrInvalidCurrency = errors.New("invalid currency. Use 'xmr'")

// Validate if the provided currency is supported
func (c Currency) Validate() (err error) {
	switch c {
	case CurrencyMonero:
		return nil
	default:
		return ErrInvalidCurrency
	}
}
