package domain

// Pagination carries page-bound list options shared by list endpoints.
//
// PageSize/PageNumber are uint32 so they convert to SQL bigint params
// (int64) without overflow risk; the HTTP layer bounds them via schema
// validation (page size 1–100, page number ≥ 1).
type Pagination struct {
	PageSize   uint32
	PageNumber uint32
}

// Offset returns the number of rows to skip before the first row of the
// page. A zero PageNumber underflows to a huge offset (empty result) —
// preserved deliberately for parity with the original implementation;
// HTTP validation makes it unreachable through the API.
func (p Pagination) Offset() int64 {
	return int64((p.PageNumber - 1) * p.PageSize)
}
