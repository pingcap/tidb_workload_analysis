package cmd

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/pingcap/index_advisor/optimizer"
	"github.com/pingcap/index_advisor/utils"
)

func TestPreCheck(t *testing.T) {
	server, err := utils.StartLocalTiDBServer("v7.1.0")
	must(err)
	db, err := optimizer.NewTiDBWhatIfOptimizer(server.DSN())
	must(err)
	err = PreCheck(db)
	if !strings.Contains(err.Error(), "your TiDB version does not support hypothetical index feature") {
		panic("should not pass")
	}
	must(server.Release())

	server, err = utils.StartLocalTiDBServer("nightly")
	must(err)
	db, err = optimizer.NewTiDBWhatIfOptimizer(server.DSN())
	must(err)
	err = PreCheck(db)
	must(err)
	must(server.Release())
}

func TestOnlineModeSimple(t *testing.T) {
	server, err := utils.StartLocalTiDBServer("nightly")
	must(err)
	defer server.Release()
	db, err := optimizer.NewTiDBWhatIfOptimizer(server.DSN())
	must(err)

	must(db.Execute(`use test`))
	must(db.Execute(`create table t (a int, b int, c int)`))
	must(db.Execute(`select a from t where a=1`))
	must(db.Execute(`select a, b from t where a=1 and b=1`))

	_, _, _, err = adviseOnlineMode(adviseOnlineCmdOpt{
		maxNumIndexes:           5,
		maxIndexWidth:           3,
		dsn:                     server.DSN(),
		output:                  "",
		logLevel:                "info",
		querySchemas:            []string{},
		queryExecTimeThreshold:  0,
		queryExecCountThreshold: 0,
		queryPath:               "",
	})
	mustTrue(strings.Contains(err.Error(), "query-schemas is not specified"), err.Error())

	_, _, _, err = adviseOnlineMode(adviseOnlineCmdOpt{
		maxNumIndexes:           5,
		maxIndexWidth:           3,
		dsn:                     server.DSN(),
		output:                  "",
		logLevel:                "info",
		querySchemas:            []string{"mysql"},
		queryExecTimeThreshold:  0,
		queryExecCountThreshold: 0,
		queryPath:               "",
	})
	mustTrue(strings.Contains(err.Error(), "no queries are found"), err.Error())

	result, _, _, err := adviseOnlineMode(adviseOnlineCmdOpt{
		maxNumIndexes:           1,
		maxIndexWidth:           3,
		dsn:                     server.DSN(),
		output:                  "",
		logLevel:                "info",
		querySchemas:            []string{"test"},
		queryExecTimeThreshold:  0,
		queryExecCountThreshold: 0,
		queryPath:               "",
	})
	must(err)
	checkAdviseResult(result, []string{"CREATE INDEX idx_a_b ON test.t (a, b)"})
}

func TestOnlineModeCrossDBQuery(t *testing.T) {
	server, err := utils.StartLocalTiDBServer("nightly")
	must(err)
	defer server.Release()
	db, err := optimizer.NewTiDBWhatIfOptimizer(server.DSN())
	must(err)

	must(db.Execute(`create database db1`))
	must(db.Execute(`use db1`))
	must(db.Execute(`create table t1 (a int, b int, c int)`))
	must(db.Execute(`create database db2`))
	must(db.Execute(`use db2`))
	must(db.Execute(`create table t2 (a int, b int, c int)`))

	must(db.Execute(`use db1`))
	must(db.Execute(`select a from t1 where a=1 and b=1`))

	must(db.Execute(`use db2`))
	must(db.Execute(`select a from t2 where a=1 and b=1`))
	must(db.Execute(`select * from t2, db1.t1 where t1.a=t2.a and t2.b=1`))

	result, _, _, err := adviseOnlineMode(adviseOnlineCmdOpt{
		maxNumIndexes:           5,
		maxIndexWidth:           3,
		dsn:                     server.DSN(),
		output:                  "",
		logLevel:                "info",
		querySchemas:            []string{"db1"},
		queryExecTimeThreshold:  0,
		queryExecCountThreshold: 0,
		queryPath:               "",
	})
	must(err)
	checkAdviseResult(result, []string{"CREATE INDEX idx_a_b ON db1.t1 (a, b)"})

	result, _, _, err = adviseOnlineMode(adviseOnlineCmdOpt{
		maxNumIndexes:           5,
		maxIndexWidth:           3,
		dsn:                     server.DSN(),
		output:                  "",
		logLevel:                "info",
		querySchemas:            []string{"db1", "db2"},
		queryExecTimeThreshold:  0,
		queryExecCountThreshold: 0,
		queryPath:               "",
	})
	must(err)
	checkAdviseResult(result, []string{"CREATE INDEX idx_a_b ON db1.t1 (a, b)",
		"CREATE INDEX idx_b ON db2.t2 (b)"})
}

func checkAdviseResult(result utils.Set[utils.Index], expected []string) {
	var got []string
	for _, r := range result.ToList() {
		got = append(got, r.DDL())
	}
	sort.Strings(got)
	sort.Strings(expected)
	gotStr := strings.Join(got, "; ")
	expStr := strings.Join(expected, "; ")
	mustTrue(gotStr == expStr, fmt.Sprintf("got %s, expected %s", gotStr, expStr))
}
