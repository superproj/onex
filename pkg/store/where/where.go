package where

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// defaultLimit defines the default limit for pagination.
	defaultLimit = -1
)

// Tenant represents a tenant with a key and a function to retrieve its value.
type Tenant struct {
	Key       string                           // The key associated with the tenant
	ValueFunc func(ctx context.Context) string // Function to retrieve the tenant's value based on the context
}

type Where interface {
	Where(db *gorm.DB) *gorm.DB
}

// WhereOption defines a function type that modifies WhereOptions.
type WhereOption func(*WhereOptions)

// WhereOptions holds the options for GORM's Where query conditions.
type WhereOptions struct {
	// Offset defines the starting point for pagination.
	// +optional
	Offset int
	// Limit defines the maximum number of results to return.
	// +optional
	Limit int
	// Filters contains key-value pairs for filtering records.
	Filters map[any]any
	// Clauses contains custom clauses to be appended to the query.
	Clauses []clause.Expression
}

// tenant holds the registered tenant instance.
var registeredTenant Tenant

// WithOffset initializes the Offset field in WhereOptions with the given offset value.
func WithOffset(offset int64) WhereOption {
	return func(whr *WhereOptions) {
		if offset < 0 {
			offset = 0
		}
		whr.Offset = int(offset)
	}
}

// WithLimit initializes the Limit field in WhereOptions with the given limit value.
func WithLimit(limit int64) WhereOption {
	return func(whr *WhereOptions) {
		if limit <= 0 {
			limit = defaultLimit
		}
		whr.Limit = int(limit)
	}
}

// WithPage is a sugar function to convert page and pageSize into limit and offset in WhereOptions.
// This function is commonly used in business logic to facilitate pagination.
func WithPage(page int, pageSize int) WhereOption {
	return func(whr *WhereOptions) {
		if page == 0 {
			page = 1
		}
		if pageSize == 0 {
			pageSize = defaultLimit
		}

		whr.Offset = (page - 1) * pageSize
		whr.Limit = pageSize
	}
}

// WithFilter initializes the Filters field in WhereOptions with the given filter criteria.
func WithFilter(filter map[any]any) WhereOption {
	return func(whr *WhereOptions) {
		whr.Filters = filter
	}
}

// WithClauses appends clauses to the Clauses field in WhereOptions.
func WithClauses(conds ...clause.Expression) WhereOption {
	return func(whr *WhereOptions) {
		whr.Clauses = append(whr.Clauses, conds...)
	}
}

// NewWhere constructs a new WhereOptions object, applying the given where options.
func NewWhere(opts ...WhereOption) *WhereOptions {
	whr := &WhereOptions{
		Offset:  0,
		Limit:   defaultLimit,
		Filters: map[any]any{},
		Clauses: make([]clause.Expression, 0),
	}

	for _, opt := range opts {
		opt(whr) // Apply each WhereOption to the opts.
	}

	return whr
}

// O sets the offset for the query.
func (whr *WhereOptions) O(offset int) *WhereOptions {
	if offset < 0 {
		offset = 0
	}
	whr.Offset = offset
	return whr
}

// L sets the limit for the query.
func (whr *WhereOptions) L(limit int) *WhereOptions {
	if limit <= 0 {
		limit = defaultLimit // Ensure defaultLimit is defined elsewhere
	}
	whr.Limit = limit
	return whr
}

// P sets the pagination based on the page number and page size.
func (whr *WhereOptions) P(page int, pageSize int) *WhereOptions {
	if page < 1 {
		page = 1 // Ensure page is at least 1
	}
	if pageSize <= 0 {
		pageSize = defaultLimit // Ensure defaultLimit is defined elsewhere
	}
	whr.Offset = (page - 1) * pageSize
	whr.Limit = pageSize
	return whr
}

// C adds conditions to the query.
func (whr *WhereOptions) C(conds ...clause.Expression) *WhereOptions {
	whr.Clauses = append(whr.Clauses, conds...)
	return whr
}

// T retrieves the value associated with the registered tenant using the provided context.
func (whr *WhereOptions) T(ctx context.Context) *WhereOptions {
	whr.F(registeredTenant.Key, registeredTenant.ValueFunc(ctx))
	return whr
}

// F adds filters to the query.
func (whr *WhereOptions) F(kvs ...any) *WhereOptions {
	if len(kvs)%2 != 0 {
		// Handle error: uneven number of key-value pairs
		return whr
	}

	for i := 0; i < len(kvs); i += 2 {
		key := kvs[i]
		value := kvs[i+1]
		whr.Filters[key] = value
	}

	return whr
}

// String returns a JSON representation of the WhereOptions.
func (whr *WhereOptions) String() string {
	jsonBytes, _ := json.Marshal(whr)
	return string(jsonBytes)
}

// Where applies the filters and clauses to the given gorm.DB instance.
func (whr *WhereOptions) Where(db *gorm.DB) *gorm.DB {
	return db.Where(whr.Filters).Clauses(whr.Clauses...).Offset(whr.Offset).Limit(whr.Limit)
}

// O is a convenience function to create a new WhereOptions with offset.
func O(offset int) *WhereOptions {
	return NewWhere().O(offset)
}

// L is a convenience function to create a new WhereOptions with limit.
func L(limit int) *WhereOptions {
	return NewWhere().L(limit)
}

// P is a convenience function to create a new WhereOptions with page number and page size.
func P(page int, pageSize int) *WhereOptions {
	return NewWhere().P(page, pageSize)
}

// C is a convenience function to create a new WhereOptions with conditions.
func C(conds ...clause.Expression) *WhereOptions {
	return NewWhere().C(conds...)
}

// T is a convenience function to create a new WhereOptions with tenant.
func T(ctx context.Context) *WhereOptions {
	return NewWhere().F(registeredTenant.Key, registeredTenant.ValueFunc(ctx))
}

// F is a convenience function to create a new WhereOptions with filters.
func F(kvs ...any) *WhereOptions {
	return NewWhere().F(kvs...)
}

// RegisterTenant registers a new tenant with the specified key and value function.
func RegisterTenant(key string, valueFunc func(context.Context) string) {
	registeredTenant = Tenant{
		Key:       key,
		ValueFunc: valueFunc,
	}
}
