package model

type Page[T any] struct {
	Content       []T   `json:"content"`
	Page          int   `json:"page"`
	Size          int   `json:"size"`
	TotalElements int64 `json:"total_elements"`
	TotalPages    int   `json:"total_pages"`
	First         bool  `json:"first"`
	Last          bool  `json:"last"`
}

func NewPage[T any](content []T, page int, size int, totalElements int64) Page[T] {
	totalPages := 0
	if totalElements > 0 {
		totalPages = int((totalElements + int64(size) - 1) / int64(size))
	}

	return Page[T]{
		Content:       content,
		Page:          page,
		Size:          size,
		TotalElements: totalElements,
		TotalPages:    totalPages,
		First:         page == 0,
		Last:          totalPages == 0 || page >= totalPages-1,
	}
}
