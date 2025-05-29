package utils

import (
	"context"
	"time"
)

const DefaultTimeout = 10_000 * time.Millisecond

func NewContext() (ctx context.Context, cancel func()) {
	return context.WithTimeout(context.TODO(), DefaultTimeout)
}
