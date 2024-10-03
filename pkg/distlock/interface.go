package distlock

import (
	"context"
	"time"
)

type Locker interface {
	Obtain(ctx context.Context) error
	Renew() error
	SetTTL(duration time.Duration) error
	Release() error
}
