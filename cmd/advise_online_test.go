package cmd

import (
	"fmt"
	"sort"
	"testing"

	"github.com/qw4990/index_advisor/optimizer"
	"github.com/qw4990/index_advisor/utils"
)

func TestReadQueries(t *testing.T) {
	server, err := utils.StartLocalTiDBServer("nightly")
	must(err)
	defer server.Release()
	dsn := server.DSN()
	db, err := optimizer.NewTiDBWhatIfOptimizer(dsn)
	must(err)

	must(db.Execute(`drop database if exists read_queries_test`))
	must(db.Execute(`create database read_queries_test`))
	defer func() {
		must(db.Execute(`drop database if exists read_queries_test`))
	}()
	must(db.Execute(`use read_queries_test`))
	must(db.Execute(`create table t1 (a int)`))
	must(db.Execute(`insert into t1 values (1)`))

	queries := []string{
		`select * from t1`,
		`select * from t1 where a in (1, 2, 3)`,
		`select * from t1 where a > 10`,

		`select a+1 from t1`,
		`select a+1 from t1`,

		`select a from t1`,
		`select a from t1`,
		`select a from t1`,

		`select a, sleep(1) from t1`,
	}
	for _, q := range queries {
		must(db.Execute(q))
	}
	// Queries below should be ignored
	must(db.Execute(`select * from information_schema.statements_summary`))
	must(db.Execute(`use mysql`))
	must(db.Execute(`select * from bind_info`))

	check := func(expected []string, opt adviseOnlineCmdOpt) {
		sqls, _ := readQueriesFromStatementSummary(db, opt.querySchemas, opt.queryExecTimeThreshold, opt.queryExecCountThreshold)
		sqls, _ = filterSQLAccessingSystemTables(sqls)
		if sqls.Size() != len(expected) {
			t.Fatalf("expect %+v, got %+v", expected, sqls)
		}
		for _, q := range expected {
			if !sqls.Contains(utils.Query{Text: q}) {
				t.Fatalf("expect %+v, got %+v", expected, sqls)
			}
		}
	}

	check([]string{`select * from t1`, `select * from t1 where a in (1, 2, 3)`, `select a+1 from t1`,
		`select * from t1 where a > 10`, `select a from t1`, `select a, sleep(1) from t1`},
		adviseOnlineCmdOpt{querySchemas: []string{"read_queries_test"}})

	check([]string{`select a+1 from t1`, `select a from t1`},
		adviseOnlineCmdOpt{querySchemas: []string{"read_queries_test"}, queryExecCountThreshold: 2})
	check([]string{`select a from t1`},
		adviseOnlineCmdOpt{querySchemas: []string{"read_queries_test"}, queryExecCountThreshold: 3})
}

func TestReadTableSchemas(t *testing.T) {
	server, err := utils.StartLocalTiDBServer("nightly")
	must(err)
	defer server.Release()
	dsn := server.DSN()
	db, err := optimizer.NewTiDBWhatIfOptimizer(dsn)
	must(err)

	must(db.Execute(`drop database if exists read_table_name_test`))
	must(db.Execute(`create database read_table_name_test`))
	defer func() {
		must(db.Execute(`drop database if exists read_table_name_test`))
	}()
	must(db.Execute(`use read_table_name_test`))
	must(db.Execute(`create table t1 (a int)`))
	must(db.Execute(`create table t2 (a int)`))
	must(db.Execute(`create table t3 (a int)`))

	schemas, _ := readTableSchemas(db, []string{`read_table_name_test`})
	if !schemas.ContainsKey("read_table_name_test.t1") ||
		!schemas.ContainsKey("read_table_name_test.t2") ||
		!schemas.ContainsKey("read_table_name_test.t3") {
		t.Fatalf("expect t1, t2, t3, got %+v", schemas)
	}
}

func TestReadTableNames(t *testing.T) {
	server, err := utils.StartLocalTiDBServer("nightly")
	must(err)
	defer server.Release()
	dsn := server.DSN()
	db, err := optimizer.NewTiDBWhatIfOptimizer(dsn)
	must(err)

	must(db.Execute(`drop database if exists read_table_name_test`))
	must(db.Execute(`create database read_table_name_test`))
	defer func() {
		must(db.Execute(`drop database if exists read_table_name_test`))
	}()
	must(db.Execute(`use read_table_name_test`))
	must(db.Execute(`create table t1 (a int)`))
	must(db.Execute(`create table t2 (a int)`))
	must(db.Execute(`create table t3 (a int)`))

	names, _ := readTableNames(db, "read_table_name_test")
	sort.Strings(names)
	if len(names) != 3 || names[0] != "t1" || names[1] != "t2" || names[2] != "t3" {
		t.Fatalf("expect t1, t2, t3, got %+v", names)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustTrue(b bool, args ...interface{}) {
	if !b {
		panic(fmt.Sprintf("%v", args))
	}
}
