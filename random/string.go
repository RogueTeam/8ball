package random

import (
	"math/rand/v2"
)

const (
	CharsetAlphaNumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharsetDigits       = "0123456789"
	CharsetSpecial      = "!@#$%^&*()_+"
)

func String(r *rand.Rand, options string, length int) (s string) {
	rOptions := []rune(options)

	var temp = make([]rune, length)
	for index := range temp {
		temp[index] = rOptions[r.IntN(len(rOptions))]
	}
	return string(temp)
}
