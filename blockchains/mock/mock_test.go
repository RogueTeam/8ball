package mock_test

import (
	"testing"

	"anarchy.ttfm/8ball/blockchains/mock"
	"anarchy.ttfm/8ball/blockchains/testsuite"
)

func Test_Mock(t *testing.T) {
	testsuite.Test(t, mock.New(), &testsuite.MockGenerator{})
}
