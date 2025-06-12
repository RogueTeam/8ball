package mock_test

import (
	"testing"

	"anarchy.ttfm/8ball/blockchains/mock"
	"anarchy.ttfm/8ball/blockchains/testsuite"
)

type genMock struct {
}

func (g *genMock) TransferAmount() (amount uint64) {
	return 1000000
}

func Test_Mock(t *testing.T) {
	testsuite.Test(t, mock.New(), &genMock{})
}
