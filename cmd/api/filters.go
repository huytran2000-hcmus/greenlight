package main

import "huytran2000-hcmus/greenlight/internal/validator"

type Filters struct {
	Page          int
	PageSize      int
	Sort          string
	SortWhiteList []string
}

func validateFilter(v *validator.Validator, f Filters) {
	v.CheckError(f.Page > 0, "page", "must be greater than zero")
	v.CheckError(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.CheckError(f.PageSize > 0, "page_size", "must be greater than zero")
	v.CheckError(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.CheckError(validator.PermittedValue(f.Sort, f.SortWhiteList...), "sort", "invalid sort value")
}
