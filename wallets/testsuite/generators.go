package testsuite

import "anarchy.ttfm/8ball/wallets/monero"

type MockGenerator struct {
}

func (g *MockGenerator) TransferAmount() (amount uint64) {
	return 1000000
}

type MoneroGenerator struct {
}

func (g *MoneroGenerator) TransferAmount() (amount uint64) {
	const value = monero.MoneroUnit / 100 // 0.01 XMR
	return value
}
