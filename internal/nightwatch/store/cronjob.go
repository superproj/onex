package store

import (
	"context"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	genericstore "github.com/superproj/onex/pkg/store"
	"github.com/superproj/onex/pkg/store/logger/onex"
	"github.com/superproj/onex/pkg/store/where"
)

// CronJobStore defines the interface for managing cron jobs in the database.
type CronJobStore interface {
	// Create inserts a new cron job into the database.
	Create(ctx context.Context, cronJob *model.CronJobM) error

	// Update modifies an existing cron job in the database.
	Update(ctx context.Context, cronJob *model.CronJobM) error

	// Delete removes cron jobs with the specified options.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a cron job with the specified options.
	Get(ctx context.Context, opts *where.WhereOptions) (*model.CronJobM, error)

	// List returns a list of cron jobs with the specified options.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.CronJobM, error)

	CronJobExpansion
}

// CronJobExpansion defines additional methods for cronjob operations.
type CronJobExpansion interface{}

// cronJobStore implements the CronJobStore interface.
type cronJobStore struct {
	*genericstore.Store[model.CronJobM]
}

// Ensure cronJobStore implements the CronJobStore interface.
var _ CronJobStore = (*cronJobStore)(nil)

// newCronJobStore creates a new instance of cronJobStore with the provided database connection.
func newCronJobStore(ds *datastore) *cronJobStore {
	return &cronJobStore{
		Store: genericstore.NewStore[model.CronJobM](ds, onex.NewLogger()),
	}
}
