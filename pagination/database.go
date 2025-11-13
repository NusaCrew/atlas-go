package pagination

import (
	api_v1 "github.com/NusaCrew/atlas-go/example/protos/api/v1"

	sq "github.com/Masterminds/squirrel"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ApplySQLPagination(builder sq.SelectBuilder, pagination *api_v1.PaginationParam) sq.SelectBuilder {
	if pagination == nil {
		return builder
	}
	if pagination.Page == 0 {
		pagination.Page = defaultPage
	}

	if pagination.PageSize == 0 {
		pagination.PageSize = defaultPageSize
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	return builder.Offset(uint64(offset)).Limit(uint64(pagination.PageSize))
}

func ApplyMongoPagination(opts *options.FindOptions, pagination *api_v1.PaginationParam) *options.FindOptions {
	if pagination == nil {
		return opts
	}

	if pagination.Page == 0 {
		pagination.Page = defaultPage
	}

	if pagination.PageSize == 0 {
		pagination.PageSize = defaultPageSize
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	return opts.SetSkip(int64(offset)).SetLimit(int64(pagination.PageSize))
}
