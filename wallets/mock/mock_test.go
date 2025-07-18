package mock_test

import (
	"testing"

	"github.com/RogueTeam/8ball/wallets/mock"
	"github.com/RogueTeam/8ball/wallets/testsuite"
)

func Test_Mock(t *testing.T) {
	testsuite.Test(t, mock.New(mock.Config{FundsDelta: 0}), &testsuite.MockGenerator{})
}
