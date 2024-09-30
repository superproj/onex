package store

import (
	"context"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	genericstore "github.com/superproj/onex/pkg/store"
	"github.com/superproj/onex/pkg/store/logger/onex"
	"github.com/superproj/onex/pkg/store/where"
)

// JobStore defines the interface for managing jobs in the database.
type JobStore interface {
	// Create inserts a new job into the database.
	Create(ctx context.Context, job *model.JobM) error

	// Update modifies an existing job in the database.
	Update(ctx context.Context, job *model.JobM) error

	// Delete removes jobs by tenant and a list of IDs.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a job by tenant and ID.
	Get(ctx context.Context, opts *where.WhereOptions) (*model.JobM, error)

	// List returns a list of jobs with the specified options.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.JobM, error)

	JobExpansion
}

// JobExpansion defines additional methods for job operations.
type JobExpansion interface{}

// jobs implements the JobStore interface.
type jobs struct {
	*genericstore.Store[model.JobM]
}

// Ensure jobs implements the JobStore interface.
var _ JobStore = (*jobs)(nil)

// newJobs creates a new instance of jobs with the provided database connection.
func newJobs(ds *datastore) *jobs {
	return &jobs{
		Store: genericstore.NewStore[model.JobM](ds, genericstore.WithLogger[model.JobM](onex.NewLogger())),
	}
}
