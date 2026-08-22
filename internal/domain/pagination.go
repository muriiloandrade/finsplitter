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
// page. A zero PageNumber defaults to the first page, so the offset is
// always a sane non-negative value.
func (p Pagination) Offset() int64 {
	page := max(p.PageNumber, 1)
	return int64((page - 1) * p.PageSize)
}
