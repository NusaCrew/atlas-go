package pagination

import (
	"math"

	api_v1 "github.com/NusaCrew/atlas-go/example/protos/api/v1"
)

const (
	defaultPage     uint32 = 1
	defaultPageSize uint32 = 20
)

func ConstructPaginationResponse(currentPage, pageSize, totalItems uint32) *api_v1.PaginationResponse {
	if currentPage == 0 {
		currentPage = defaultPage
	}

	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	return &api_v1.PaginationResponse{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalPages:  uint32(math.Ceil(float64(totalItems) / float64(pageSize))),
		TotalItems:  totalItems,
	}
}
