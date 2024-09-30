package store

//go:generate mockgen -destination mock_store.go -package store github.com/superproj/onex/internal/nightwatch/store IStore,CronJobStore,JobStore

import (
	"context"
	"sync"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet contains providers for creating instances of the datastore struct using Google Wire.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*datastore)))

var (
	once sync.Once
	// S is a global variable that holds the initialized instance of datastore for convenient access by other packages.
	S *datastore
)

// transactionKey is an unique key used in context to store
// transaction instances to be shared between multiple operations.
type transactionKey struct{}

// IStore defines the interface for the store layer, specifying the methods that need to be implemented.
type IStore interface {
	DB(ctx context.Context) *gorm.DB
	TX(ctx context.Context, fn func(ctx context.Context) error) error
	CronJobs() CronJobStore
	Jobs() JobStore
}

// datastore is a concrete implementation of the IStore interface.
type datastore struct {
	db *gorm.DB // Database connection.
}

// Ensure datastore implements the IStore interface.
var _ IStore = (*datastore)(nil)

// NewStore creates a new instance of the datastore struct, implementing the IStore interface.
func NewStore(db *gorm.DB) *datastore {
	// Ensure S is initialized only once.
	once.Do(func() {
		S = &datastore{db}
	})

	return S
}

// DB retrieves the current database instance from the context or returns the main instance.
func (ds *datastore) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(transactionKey{}).(*gorm.DB)
	if ok {
		return tx
	}

	return ds.db
}

// TX starts a transaction using the main DB context
// and passes the transactional context to the provided function.
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return ds.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

// CronJobs returns an instance that implements the CronJobStore interface.
func (ds *datastore) CronJobs() CronJobStore {
	return newCronJobs(ds)
}

// Jobs returns an instance that implements the JobStore interface.
func (ds *datastore) Jobs() JobStore {
	return newJobs(ds)
}
