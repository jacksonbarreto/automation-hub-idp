package utils

type Pagination struct {
	Limit  int `default:"10"`
	Offset int `default:"0"`
}

func NewPagination(limit, offset int) Pagination {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{Limit: limit, Offset: offset}
}

func DefaultPagination() Pagination {
	return Pagination{Limit: 10, Offset: 0}
}
