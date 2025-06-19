package mock_test

import (
	"testing"

	"anarchy.ttfm/8ball/wallets/mock"
	"anarchy.ttfm/8ball/wallets/testsuite"
)

func Test_Mock(t *testing.T) {
	testsuite.Test(t, mock.New(mock.Config{FundsDelta: 0}), &testsuite.MockGenerator{})
}
