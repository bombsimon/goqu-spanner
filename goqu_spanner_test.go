package goquspanner

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"

	_ "github.com/googleapis/go-sql-spanner"
)

func setup() goqu.DialectWrapper {
	return goqu.Dialect(DialectName)
}

func TestSelect(t *testing.T) {
	dw := setup()

	cases := []struct {
		description string
		dataset     *goqu.SelectDataset
		expectedSQL string
	}{
		{
			description: "select star",
			dataset:     dw.From("table"),
			expectedSQL: "SELECT * FROM `table`",
		},
		{
			description: "select columns",
			dataset:     dw.From("table").Select("a", "b", "c"),
			expectedSQL: "SELECT `a`, `b`, `c` FROM `table`",
		},
		{
			description: "select literal and alias",
			dataset: dw.From("table").
				Select(goqu.L("a + b").As("sum"), goqu.C("c").As("cc")).
				Where(
					goqu.Ex{"a": 3},
					goqu.C("b").IsNotNull(),
					goqu.C("cc").Gt(2),
				),
			expectedSQL: "SELECT a + b AS `sum`, `c` AS `cc` FROM `table` WHERE ((`a` = 3) AND (`b` IS NOT NULL) AND (`cc` > 2))",
		},
		{
			description: "select count and functions",
			dataset: dw.From(goqu.T("table").As("the_table")).
				Order(goqu.C("age").Desc()).
				Select(
					goqu.COUNT("*").As("age_count"),
					goqu.MAX("age").As("max_age"),
					goqu.AVG("age").As("avg_age"),
				),
			expectedSQL: "SELECT COUNT(*) AS `age_count`, MAX(`age`) AS `max_age`, AVG(`age`) AS `avg_age` FROM `table` AS `the_table` ORDER BY `age` DESC",
		},
		{
			description: "select sub query",
			dataset: dw.From(
				dw.From("sub_table").
					Where(goqu.C("age").Gt(10)).
					As("temp_table"),
			),
			expectedSQL: "SELECT * FROM (SELECT * FROM `sub_table` WHERE (`age` > 10)) AS `temp_table`",
		},
		{
			description: "select distinct column",
			dataset: dw.From("table").
				Select(
					goqu.L("COALESCE(?, ?)", goqu.C("a"), "empty"),
				).Distinct(),
			expectedSQL: "SELECT DISTINCT COALESCE(`a`, 'empty') FROM `table`",
		},
		{
			description: "select in (not unnest)",
			dataset: dw.From("table").
				Where(goqu.C("col").In([]string{"a", "b", "c"})),
			expectedSQL: "SELECT * FROM `table` WHERE (`col` IN ('a', 'b', 'c'))",
		},
		{
			description: "select literal and identifier",
			dataset: dw.From("table").
				Where(
					goqu.L(
						"(? AND ?) OR (?)",
						goqu.I("a").Eq(1),
						goqu.I("b").Eq("b"),
						goqu.I("c").In([]string{"a", "b", "c"}),
					),
				),
			expectedSQL: "SELECT * FROM `table` WHERE ((`a` = 1) AND (`b` = 'b')) OR ((`c` IN ('a', 'b', 'c')))",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			gotSQL, _, _ := tc.dataset.ToSQL()
			assert.Equal(t, tc.expectedSQL, gotSQL)
		})
	}
}

func TestNotSupported(t *testing.T) {
	dw := setup()

	cases := []struct {
		description string
		dataset     *goqu.SelectDataset
	}{
		{
			description: "select distinct on",
			dataset: dw.From("table").
				Distinct(
					goqu.L("COALESCE(?, ?)", goqu.C("a"), "empty"),
				),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			gotSQL, _, _ := tc.dataset.ToSQL()
			assert.Empty(t, gotSQL)
		})
	}
}

func _boilerplate(t *testing.T) {
	dw := setup()

	cases := []struct {
		description string
		dataset     *goqu.SelectDataset
		expectedSQL string
	}{
		{
			description: "select star",
			dataset:     dw.From("table"),
			expectedSQL: "SELECT * FROM `table`",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			gotSQL, _, _ := tc.dataset.ToSQL()
			assert.Equal(t, tc.expectedSQL, gotSQL)
		})
	}
}
