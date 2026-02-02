package domain

import "context"

// Transactional supports running operations inside a transaction.
type Transactional interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CursorPagination is a cursor-based pagination request.
type CursorPagination struct {
	Cursor *int64
	Limit  int
}

// CursorPaginationResponse is a cursor-based pagination response.
type CursorPaginationResponse struct {
	Total  int    `json:"total"`
	Cursor *int64 `json:"cursor,omitempty"`
}
