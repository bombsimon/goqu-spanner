package goquspanner

import (
	"github.com/doug-martin/goqu/v9"
)

const dialectName = "spanner"

// DialectOptions returns the SQL options to implement a goqu dialect.
func DialectOptions() *goqu.SQLDialectOptions {
	options := goqu.DefaultDialectOptions()

	return options
}

//nolint:gochecknoinits // This is how a dialect is loaded.
func init() {
	goqu.RegisterDialect(dialectName, DialectOptions())
}
