package biz

//go:generate mockgen -destination mock_biz.go -package biz github.com/superproj/onex/internal/nightwatch/biz IBiz

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/nightwatch/biz/cronjob"
	"github.com/superproj/onex/internal/nightwatch/biz/job"
	"github.com/superproj/onex/internal/nightwatch/store"
)

// ProviderSet contains providers for creating instances of the biz struct using Google Wire.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines the interface for accessing business logic related to cron jobs and jobs.
type IBiz interface {
	CronJobs() cronjob.CronJobBiz
	Jobs() job.JobBiz
}

// biz is the concrete implementation of the IBiz interface.
type biz struct {
	ds store.IStore // Data store interface for accessing data.
}

// Ensure biz implements the IBiz interface.
var _ IBiz = (*biz)(nil)

// NewBiz creates a new instance of the biz struct, implementing the IBiz interface.
func NewBiz(ds store.IStore) *biz {
	return &biz{ds: ds}
}

// CronJobs returns the business logic interface for managing cron jobs.
func (b *biz) CronJobs() cronjob.CronJobBiz {
	return cronjob.New(b.ds)
}

// Jobs returns the business logic interface for managing jobs.
func (b *biz) Jobs() job.JobBiz {
	return job.New(b.ds)
}
