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

	// Delete removes jobs with the specified options.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a job with the specified options..
	Get(ctx context.Context, opts *where.WhereOptions) (*model.JobM, error)

	// List returns a list of jobs with the specified options.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.JobM, error)

	JobExpansion
}

// JobExpansion defines additional methods for job operations.
type JobExpansion interface{}

// jobStore implements the JobStore interface.
type jobStore struct {
	*genericstore.Store[model.JobM]
}

// Ensure jobStore implements the JobStore interface.
var _ JobStore = (*jobStore)(nil)

// newJobStore creates a new instance of jobStore with the provided database connection.
func newJobStore(ds *datastore) *jobStore {
	return &jobStore{
		Store: genericstore.NewStore[model.JobM](ds, onex.NewLogger()),
	}
}
