package testsuite

type MockGenerator struct {
}

func (g *MockGenerator) TransferAmount() (amount uint64) {
	return 1000000
}

type MoneroGenerator struct {
}

func (g *MoneroGenerator) TransferAmount() (amount uint64) {
	return 1000000000
}
