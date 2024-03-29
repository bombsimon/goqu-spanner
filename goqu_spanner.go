package goquspanner

import (
	"github.com/doug-martin/goqu/v9"
)

// DialectName is the name of the dialect.
const DialectName = "spanner"

// DialectOptions returns the SQL options to implement a goqu dialect.
func DialectOptions() *goqu.SQLDialectOptions {
	options := goqu.DefaultDialectOptions()

	options.SupportsConflictTarget = false
	options.SupportsConflictUpdateWhere = false
	options.SupportsDistinctOn = false
	options.SupportsDistinctOn = false
	options.SupportsMultipleUpdateTables = false
	options.SupportsReturn = false
	options.SupportsWindowFunction = false

	options.QuoteRune = '`'

	return options
}

//nolint:gochecknoinits // This is how a dialect is loaded.
func init() {
	goqu.RegisterDialect(DialectName, DialectOptions())
}
