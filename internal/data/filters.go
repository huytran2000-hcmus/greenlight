package data

import (
	"strings"

	"huytran2000-hcmus/greenlight/internal/validator"
)

type Filters struct {
	Page          int
	PageSize      int
	Sort          string
	SortWhiteList []string
}

func ValidateFilter(v *validator.Validator, f Filters) {
	v.CheckError(f.Page > 0, "page", "must be greater than zero")
	v.CheckError(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.CheckError(f.PageSize > 0, "page_size", "must be greater than zero")
	v.CheckError(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.CheckError(validator.PermittedValue(f.Sort, f.SortWhiteList...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	for _, val := range f.SortWhiteList {
		if f.Sort == val {
			return strings.Trim(f.Sort, "-")
		}
	}

	return ""
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}
