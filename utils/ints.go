package utils

import (
	"golang.org/x/exp/constraints"
)

func MapInt[T constraints.Integer | constraints.Float, D constraints.Integer | constraints.Float](src []T) (dst []D) {
	dst = make([]D, 0, len(src))
	for _, value := range src {
		dst = append(dst, D(value))
	}
	return dst
}
