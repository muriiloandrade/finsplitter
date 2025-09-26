package postgres

import "github.com/Masterminds/squirrel"

//nolint:gochecknoglobals // meant to be used on repositories
var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
