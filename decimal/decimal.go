package decimal

import (
	"encoding/json"
	"math/big"

	"anarchy.ttfm/8ball/wallets/monero"
)

var MoneroAsBigFloat = big.NewFloat(0).SetMode(RoundingMode).SetPrec(OperationPrec).SetInt(big.NewInt(0).SetUint64(monero.MoneroUnit))

type Decimal struct {
	Value *big.Float
}

const OperationPrec = 256

const RoundingMode = big.AwayFromZero

func (d *Decimal) FromUint64(v uint64) {
	d.Value = big.NewFloat(0).SetMode(RoundingMode).SetPrec(OperationPrec).SetInt(big.NewInt(0).SetUint64(v))
	d.Value = d.Value.Quo(d.Value, MoneroAsBigFloat)
}

func (d *Decimal) ToUint64() (v uint64) {
	var amountCopy big.Float
	amountCopy = *amountCopy.Copy(d.Value)
	asInt, _ := amountCopy.Mul(&amountCopy, MoneroAsBigFloat).Int(nil)
	return asInt.Uint64()
}

func (d *Decimal) FromString(s string) (err error) {
	d.Value, _, err = big.ParseFloat(s, 10, OperationPrec, RoundingMode)
	if err != nil {
		return err
	}
	return nil
}

var (
	_ json.Unmarshaler = (*Decimal)(nil)
	_ json.Marshaler   = (*Decimal)(nil)
)

func (d *Decimal) UnmarshalJSON(b []byte) (err error) {
	var asString string
	err = json.Unmarshal(b, &asString)
	if err != nil {
		return err
	}

	return d.FromString(asString)
}

func (d *Decimal) MarshalJSON() (b []byte, err error) {
	return []byte("\"" + d.Value.Text('f', 12) + "\""), nil
}
