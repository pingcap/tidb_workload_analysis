package advisor

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/pingcap/index_advisor/optimizer"
	"github.com/pingcap/index_advisor/utils"
)

func prepareTable(db optimizer.WhatIfOptimizer, schema, createStmt string, nRows int) {
	t, err := utils.ParseCreateTableStmt(schema, createStmt)
	must(err)

	must(db.Execute(createStmt))
	for i := 0; i < nRows; i++ {
		var values []string
		for j := 0; j < len(t.Columns); j++ {
			values = append(values, fmt.Sprintf("%v", i*100+rand.Intn(100)))
		}
		must(db.Execute(fmt.Sprintf("insert into %v values (%v)", t.TableName, strings.Join(values, ","))))
	}
	db.Execute(`analyze table ` + t.TableName)
}

func prepareTestIndexSelectionAAEnd2End(db optimizer.WhatIfOptimizer, schema string, createStmts []string, rows int) {
	must(db.Execute(`drop database if exists ` + schema))
	must(db.Execute(`create database ` + schema))
	must(db.Execute(`use ` + schema))

	for _, createStmt := range createStmts {
		prepareTable(db, schema, createStmt, rows)
	}
}

func TestIndexSelectionEnd2End(t *testing.T) {
	server, err := utils.StartLocalTiDBServer("nightly")
	must(err)
	defer server.Release()
	dsn := server.DSN()
	db, err := optimizer.NewTiDBWhatIfOptimizer(dsn)
	must(err)

	schema := "test"
	createTableStmts := []string{
		`create table t1 (a int)`,
		`create table t2 (a int, b int)`,
		`create table t3 (a int, b int, c int)`,
	}
	prepareTestIndexSelectionAAEnd2End(db, schema, createTableStmts, 3000)

	type aaCase struct {
		queries []string
		param   Parameter
		result  []string
	}
	cases := []aaCase{
		// single-table cases
		// zero-predicate cases
		{[]string{`select * from t1`}, Parameter{1, 3},
			[]string{}}, // no index can help
		// TODO: cannot pass this case now since `a` is not considered as an indexable column.
		//{[]string{`select a from t1`}, Parameter{1, 3},
		//	[]string{"test.t1(a)"}}, // idx(a) can help decrease the scan cost.
		{[]string{`select a from t1 order by a`}, Parameter{1, 3},
			[]string{"test.t1(a)"}}, // idx(a) can help decrease the scan cost.
		{[]string{`select a from t1 group by a`}, Parameter{1, 3},
			[]string{"test.t1(a)"}}, // idx(a) can help decrease the scan cost.

		// 	single-predicate cases
		{[]string{`select * from t1 where a=1`}, Parameter{1, 3}, []string{"test.t1(a)"}},
		{[]string{`select * from t1 where a=1`}, Parameter{5, 3},
			[]string{"test.t1(a)"}}, // only 1 index should be generated even if it asks for 5.
		{[]string{`select * from t1 where a<50`}, Parameter{1, 3}, []string{"test.t1(a)"}},
		{[]string{`select * from t1 where a in (1, 2, 3, 4, 5)`}, Parameter{1, 3}, []string{"test.t1(a)"}},
		{[]string{`select * from t1 where a=1 order by a`}, Parameter{1, 3}, []string{"test.t1(a)"}},
		{[]string{`select * from t2 where a=1 order by b`}, Parameter{1, 3}, []string{"test.t2(a,b)"}},
		{[]string{`select * from t2 where a in (1, 2, 3) order by b`}, Parameter{1, 3}, []string{"test.t2(a,b)"}},
		{[]string{`select * from t2 where a < 20 order by b`}, Parameter{1, 3}, []string{"test.t2(a,b)"}},
		// TODO: should be t(b, a)
		{[]string{`select * from t2 where a > 20 order by b`}, Parameter{1, 3}, []string{"test.t2(a,b)"}},

		// multi-predicate cases
		{[]string{`select * from t2 where a=1 and b=1`}, Parameter{1, 3}, []string{"test.t2(a,b)"}},
		{[]string{`select * from t2 where a=1 and b=1`}, Parameter{2, 3}, []string{"test.t2(a,b)"}},
		{[]string{`select * from t2 where a=1 and b=1`}, Parameter{3, 3}, []string{"test.t2(a,b)"}},
		{[]string{`select * from t2 where a<1 and b=1`}, Parameter{1, 3}, []string{"test.t2(b,a)"}},
		{[]string{`select * from t2 where a<1 and b=1`}, Parameter{2, 3}, []string{"test.t2(b,a)"}},
		{[]string{`select * from t2 where a<1 and b=1`}, Parameter{1, 1}, []string{"test.t2(b)"}},
		{[]string{`select * from t2 where a=1 or b=1`}, Parameter{1, 1}, []string{"test.t2(a)"}},
		{[]string{`select * from t2 where a=1 or b=1`}, Parameter{1, 3}, []string{"test.t2(a,b)"}},

		// multi-queries cases
		{[]string{`select * from t1 where a=1`, `select * from t2 where a=1`}, Parameter{1, 3}, []string{"test.t1(a)"}},
		{[]string{`select * from t1 where a>1`, `select * from t2 where a=1`}, Parameter{1, 3}, []string{"test.t2(a)"}},
		{[]string{`select * from t1 where a=1`, `select * from t2 where a=1`}, Parameter{2, 3}, []string{"test.t1(a)", "test.t2(a)"}},
		{[]string{`select * from t3 where a=1`, `select * from t3 where a=2`, `select * from t3 where b=1`}, Parameter{1, 3}, []string{"test.t3(a)"}},
		{[]string{`select * from t3 where a=1`, `select * from t3 where a=2`, `select * from t3 where b=1`}, Parameter{2, 3}, []string{"test.t3(a)", "test.t3(b)"}},
		{[]string{`select * from t3 where a=1`, `select * from t3 where a=2`, `select * from t3 where b=1 and a=3`}, Parameter{1, 3}, []string{"test.t3(a,b)"}},
		{[]string{`select * from t3 where a=1`, `select * from t3 where a=2`, `select * from t3 where b=1 and a=3`}, Parameter{2, 3}, []string{"test.t3(a,b)"}},
		{[]string{`select * from t2 where a=1 and b=1`, `select * from t3 where a=1 and b=1`}, Parameter{1, 3}, []string{"test.t3(a,b)"}},
		{[]string{`select * from t2 where a=1 and b=1`, `select * from t3 where a=1 and b=1`}, Parameter{2, 3}, []string{"test.t2(a,b)", "test.t3(a,b)"}},
		//{[]string{`select * from t2 where a>1 and b=1`, `select * from t3 where a>1 and b=1`}, Parameter{1, 3}, []string{"test.t2(a,b)"}},
		//{[]string{`select * from t2 where a>1 and b=1`, `select * from t3 where a>1 and b=1`}, Parameter{2, 3}, []string{"test.t2(b,a)", "test.t3(b,a)"}},

		// index merge cases
		{[]string{`select * from t2 where a=1 or b=1`}, Parameter{2, 3}, []string{"test.t2(a)", "test.t2(b,a)"}},
		{[]string{`select * from t3 where a=1 or b=1 or c=1`}, Parameter{3, 3}, []string{"test.t3(a)", "test.t3(b)", "test.t3(c)"}},

		// cover-index cases
		{[]string{`select a from t1`}, Parameter{1, 3}, []string{"test.t1(a)"}},
		{[]string{`select a, b from t3`}, Parameter{1, 3}, []string{"test.t3(a,b)"}},
		{[]string{`select c, a, b from t3`}, Parameter{1, 3}, []string{"test.t3(a,b,c)"}},
		{[]string{`select a from t3 where b=1`}, Parameter{1, 3}, []string{"test.t3(b,a)"}},
		{[]string{`select a, c from t3 where b=1`}, Parameter{1, 3}, []string{"test.t3(b,a,c)"}},
		{[]string{`select a from t3 where b=1 and c=1`}, Parameter{1, 3}, []string{"test.t3(b,c,a)"}},
	}

	for i, c := range cases {
		workload, err := utils.CreateWorkloadFromRawStmt(schema, createTableStmts, c.queries)
		must(err)
		result, err := IndexAdvise(db, workload, c.param)
		must(err)

		var resultKeys []string
		for _, r := range result.ToList() {
			resultKeys = append(resultKeys, r.Key())
		}
		sort.Strings(resultKeys)
		sort.Strings(c.result)

		expected := strings.Join(c.result, ",")
		actual := strings.Join(resultKeys, ",")
		if expected != actual {
			t.Errorf("case: %v, expected: %v, actual: %v, query: %v", i, expected, actual, c.queries)
			break
		}
	}
}
