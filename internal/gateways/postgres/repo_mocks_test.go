package postgres

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ---------------------------------------------------------------------------
// mockRow — implements pgx.Row by returning pre-set values/errors.
// ---------------------------------------------------------------------------

type mockRow struct {
	values []any
	err    error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	for i, d := range dest {
		if i >= len(m.values) {
			break
		}
		reflectAssign(d, m.values[i])
	}
	return nil
}

// ---------------------------------------------------------------------------
// mockRows — implements pgx.Rows by iterating over pre-set row slices.
// ---------------------------------------------------------------------------

type mockRows struct {
	pgx.Rows // embedded nil — satisfies interface at compile time

	rows    [][]any
	cur     int
	closed  bool
	scanErr error // returned by Scan, if set
	rowsErr error // returned by Err(), if set
}

func (m *mockRows) Close() {
	m.closed = true
}

func (m *mockRows) Err() error {
	return m.rowsErr
}

func (m *mockRows) Next() bool {
	m.cur++
	return m.cur <= len(m.rows)
}

func (m *mockRows) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	idx := m.cur - 1
	if idx < 0 || idx >= len(m.rows) {
		return pgx.ErrNoRows
	}
	row := m.rows[idx]
	for i, d := range dest {
		if i >= len(row) {
			break
		}
		reflectAssign(d, row[i])
	}
	return nil
}

// ---------------------------------------------------------------------------
// mockQuerier — implements both sqlc.DBTX and the local querier interface
// ---------------------------------------------------------------------------

type mockQuerier struct {
	execFunc    func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	queryFunc   func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFn  func(ctx context.Context, sql string, args ...any) pgx.Row
	sendBatchFn func(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func (m *mockQuerier) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, arguments...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return &mockRows{}, nil
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFn != nil {
		return m.queryRowFn(ctx, sql, args...)
	}
	return &mockRow{}
}

func (m *mockQuerier) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	if m.sendBatchFn != nil {
		return m.sendBatchFn(ctx, b)
	}
	return nil
}

// Compile-time check.
var _ querier = (*mockQuerier)(nil)

// ---------------------------------------------------------------------------
// reflectAssign copies src into dest (dest must be a non-nil pointer).
// ---------------------------------------------------------------------------

func reflectAssign(dest, src any) {
	dv := reflect.ValueOf(dest)
	sv := reflect.ValueOf(src)
	if dv.Kind() == reflect.Pointer && !dv.IsNil() {
		if sv.Type().AssignableTo(dv.Elem().Type()) {
			dv.Elem().Set(sv)
		}
	}
}
