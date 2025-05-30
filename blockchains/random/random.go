package random

import (
	crand "crypto/rand"
	"math/rand/v2"
)

var (
	PseudoRand = rand.New(rand.NewPCG(0xFF_FF_FF_FF, 0xAA_BB_CC_DD))
)

func CryptoRand() (r *rand.Rand) {
	var seed [32]byte
	crand.Reader.Read(seed[:])
	return rand.New(rand.NewChaCha8(seed))
}
