package uow

import "context"

// UnitOfWork defines the interface for managing transactions
// This abstraction allows services to work with any database (SQL, MongoDB, etc.)
type UnitOfWork interface {
	// StartContext begins a new transaction and returns a context containing it
	StartContext(ctx context.Context) context.Context

	// DeferErrorContext commits or rolls back based on the error
	// If err is nil, commits the transaction; otherwise rolls back
	DeferErrorContext(ctx context.Context, err error) error
}
