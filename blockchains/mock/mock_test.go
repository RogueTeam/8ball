package mock_test

import (
	"testing"

	"anarchy.ttfm.onion/gateway/blockchains/mock"
	"anarchy.ttfm.onion/gateway/blockchains/testsuite"
)

type genMock struct {
}

func (g *genMock) TransferAmount() (amount uint64) {
	return 1000000
}

func Test_Mock(t *testing.T) {
	testsuite.Test(t, mock.New(), &genMock{})
}
