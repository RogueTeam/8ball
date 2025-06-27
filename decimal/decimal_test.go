package decimal_test

import (
	"encoding/json"
	"testing"

	"anarchy.ttfm/8ball/decimal"
	"anarchy.ttfm/8ball/wallets/monero"
	"github.com/stretchr/testify/assert"
)

func Test_FloatIntegration(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		type Test struct {
			Reference string
			Expect    uint64
		}
		tests := []Test{
			{
				Reference: `10.000000000001`,
				Expect:    10*monero.MoneroUnit + 1,
			},
			{
				Reference: `0`,
				Expect:    0,
			},
			{
				Reference: `0.0`,
				Expect:    0,
			},
			{
				Reference: `1`,
				Expect:    1 * monero.MoneroUnit,
			},
			{
				Reference: `1.0`,
				Expect:    1 * monero.MoneroUnit,
			},
			{
				Reference: `1.000000000000`,
				Expect:    1 * monero.MoneroUnit,
			},
			{
				Reference: `0.000000000001`,
				Expect:    1,
			},
			{
				Reference: `0.000000000010`,
				Expect:    10,
			},
			{
				Reference: `0.1`,
				Expect:    100000000000,
			},
			{
				Reference: `0.123456789012`,
				Expect:    123456789012,
			},
			{
				Reference: `100`,
				Expect:    100 * monero.MoneroUnit,
			},
			{
				Reference: `123.456`,
				Expect:    123_456_000_000_000,
			},
			{
				Reference: `5.5`,
				Expect:    5500000000000,
			},
			{
				Reference: `5.000000000005`,
				Expect:    5*monero.MoneroUnit + 5,
			},
			{
				Reference: `0.000000000000`,
				Expect:    0,
			},
			{
				Reference: `0.5`,
				Expect:    500000000000,
			},
			{
				Reference: `0.000000000000`,
				Expect:    0,
			},
			{
				Reference: `1.000000000001`,
				Expect:    1*monero.MoneroUnit + 1,
			},
			{
				Reference: `25.000000000000`,
				Expect:    25 * monero.MoneroUnit,
			},
			{
				Reference: `25.123456789012`,
				Expect:    25*monero.MoneroUnit + 123456789012,
			},
			{
				Reference: `0.000000000000`,
				Expect:    0,
			},
			{
				Reference: `12345.6789`,
				Expect:    12345_678_900_000_000,
			},
			{
				Reference: `0.000000000000`,
				Expect:    0,
			},
		}
		for _, test := range tests {
			name, _ := json.Marshal(test)
			t.Run(string(name), func(t *testing.T) {
				assertions := assert.New(t)

				var value decimal.Decimal
				err := value.FromString(test.Reference)
				assertions.Nil(err, "failed to convert from string")

				srcBytes, err := json.Marshal(value)
				assertions.Nil(err, "failed to marshal src")
				t.Log("Reconverted", string(srcBytes))

				var final decimal.Decimal
				final.FromUint64(value.ToUint64())
				assertions.Equal(value.ToUint64(), final.ToUint64(), "not equal")
			})
		}
	})
}
