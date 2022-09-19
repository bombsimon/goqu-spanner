# goqu-spanner

This is a [`goqu`][goqu] implementation for [Spanner].

**THIS IS A WORK-IN-PROGRESS, DON'T USE IT IN PRODUCTION YET.**.

## Usage

Use like any other goqu driver, see [`examples`][examples] for details.

```go
package main

import (
    "database/sql"

    "github.com/doug-martin/goqu/v9"

    // Load our custom dialect.
    "github.com/bombsimon/goquspanner"
    _ "github.com/googleapis/go-sql-spanner"
)

func main() {
    spannerDB, err := sql.Open(goquspanner.DialectName, emulatorDBPath)
    if err != nil {
        panic(err)
    }

    dialect := goqu.Dialect(goquspanner.DialectName)
    db := dialect.DB(spannerDB)

  _ = db
}
```

## References

- [goqu dialect]
- [Spanner query syntax]

  [examples]: ./examples
  [go-sql-spanner]: https://github.com/googleapis/go-sql-spanner
  [goqu dialect]: https://github.com/doug-martin/goqu/blob/master/docs/dialect.md
  [goqu]: https://github.com/doug-martin/goqu
  [spanner query syntax]: https://cloud.google.com/spanner/docs/reference/standard-sql/query-syntax
  [spanner]: https://cloud.google.com/spanner
