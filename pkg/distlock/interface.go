package distlock

import "context"

type Locker interface {
	Obtain(ctx context.Context) error
	Renew() error
	SetTTL(duration time.Duration) error
	Release() error
}
