package utils

import (
	"context"
	"time"
)

const DefaultTimeout = time.Minute

func NewContext() (ctx context.Context, cancel func()) {
	return context.WithTimeout(context.TODO(), DefaultTimeout)
}
