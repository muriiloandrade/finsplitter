package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination_Offset(t *testing.T) {
	tests := []struct {
		name       string
		pagination Pagination
		want       int64
	}{
		{
			name:       "first page has zero offset",
			pagination: Pagination{PageSize: 10, PageNumber: 1},
			want:       0,
		},
		{
			name:       "second page skips one page size",
			pagination: Pagination{PageSize: 10, PageNumber: 2},
			want:       10,
		},
		{
			name:       "arbitrary page skips preceding rows",
			pagination: Pagination{PageSize: 25, PageNumber: 7},
			want:       150,
		},
		{
			name:       "zero page number underflows to a huge offset returning empty results",
			pagination: Pagination{PageSize: 10, PageNumber: 0},
			// (0-1)*10 wraps mod 2^32 → MaxUint32-9; documents the preserved quirk
			want: int64(uint32(0xFFFFFFFF - 9)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.pagination.Offset())
		})
	}
}
