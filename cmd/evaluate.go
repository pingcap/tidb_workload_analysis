package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/pingcap/index_advisor/optimizer"
	"github.com/pingcap/index_advisor/utils"
	"github.com/spf13/cobra"
)

type evaluateCmdOpt struct {
	dsn          string
	analyze      bool
	qWhiteList   string
	qBlackList   string
	queryPath    string
	indexDirPath string
	output       string
}

func NewEvaluateCmd() *cobra.Command {
	var opt evaluateCmdOpt
	cmd := &cobra.Command{
		Use:    "evaluate",
		Short:  "exec all queries in the specified workload (only for test)",
		Long:   `exec all queries in the specified workload and collect their plans and execution times (only for test)`,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, dbName := utils.GetDBNameFromDSN(opt.dsn)
			if dbName == "" {
				return fmt.Errorf("invalid dsn: %s, no database name", opt.dsn)
			}

			queries, err := utils.LoadQueries(dbName, opt.queryPath)
			if err != nil {
				return err
			}

			if opt.qWhiteList != "" || opt.qBlackList != "" {
				queries = utils.FilterQueries(queries, strings.Split(opt.qWhiteList, ","), strings.Split(opt.qBlackList, ","))
			}

			db, err := optimizer.NewTiDBWhatIfOptimizer(opt.dsn)
			if err != nil {
				return err
			}

			sqls := queries.ToList()
			sort.Slice(sqls, func(i, j int) bool {
				return sqls[i].Alias < sqls[j].Alias
			})

			if opt.indexDirPath == "" {
				if opt.analyze {
					return explainQueries(db, queries)
				} else {
					return executeQueries(db, queries, opt.output)
				}
			}

			indexFiles, err := os.ReadDir(opt.indexDirPath)
			if err != nil {
				return err
			}
			for _, indexFile := range indexFiles {
				if !strings.HasSuffix(indexFile.Name(), ".sql") {
					continue
				}
				utils.Infof("execute queries with index %s", indexFile.Name())
				indexConfPath := path.Join(opt.indexDirPath, indexFile.Name())
				if err := executeQueriesWithIndexes(db, queries, indexConfPath, opt.output); err != nil {
					return err
				}
				utils.Infof("finish executing queries with index %s", indexFile.Name())
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opt.dsn, "dsn", "root:@tcp(127.0.0.1:4000)/test", "dsn")
	cmd.Flags().BoolVar(&opt.analyze, "analyze", true, "whether to use `explain analyze`")
	cmd.Flags().StringVar(&opt.queryPath, "query-path", "", "")
	cmd.Flags().StringVar(&opt.indexDirPath, "index-dir", "", "")
	cmd.Flags().StringVar(&opt.qWhiteList, "query-white-list", "", "queries to consider, e.g. 'q1,q2,q6'")
	cmd.Flags().StringVar(&opt.qBlackList, "query-black-list", "", "queries to ignore, e.g. 'q5,q12'")
	cmd.Flags().StringVar(&opt.output, "output", "", "output directory to save the result")
	return cmd
}

func executeQueriesWithIndexes(db optimizer.WhatIfOptimizer, queries utils.Set[utils.Query], indexConfPath, savePath string) error {
	// load indexes from indexConfPath into the cluster
	stmts, err := utils.ParseStmtsFromFile(indexConfPath)
	if err != nil {
		return err
	}
	indexes := utils.NewSet[utils.Index]()
	tableNames := utils.NewSet[utils.TableName]()
	for _, stmt := range stmts {
		index, err := utils.ParseCreateIndexStmt(stmt)
		if err != nil {
			return err
		}
		indexes.Add(index)
		tableNames.Add(utils.TableName{SchemaName: index.SchemaName, TableName: index.TableName})
	}
	for _, index := range indexes.ToList() {
		utils.Infof("execute: %s", index.DDL())
		if err := db.Execute(index.DDL()); err != nil {
			return err
		}
	}
	for _, t := range tableNames.ToList() {
		analyzeStmt := fmt.Sprintf("ANALYZE TABLE %s.%s", t.SchemaName, t.TableName)
		utils.Infof("execute: %s", analyzeStmt)
		if err := db.Execute(analyzeStmt); err != nil {
			return err
		}
	}

	// run queries
	baseName := path.Base(indexConfPath)
	baseName = strings.TrimSuffix(baseName, path.Ext(baseName))
	if err := executeQueries(db, queries, path.Join(savePath, baseName)); err != nil {
		return err
	}

	// remove indexes in indexConfPath
	for _, index := range indexes.ToList() {
		dropStmt := fmt.Sprintf("DROP INDEX %s ON %s.%s", index.IndexName, index.SchemaName, index.TableName)
		utils.Infof("execute: %s", dropStmt)
		if err := db.Execute(dropStmt); err != nil {
			return err
		}
	}
	return nil
}

func explainQueries(db optimizer.WhatIfOptimizer, queries utils.Set[utils.Query]) error {
	queryList := queries.ToList()
	var totCost float64
	type qCost struct {
		Alias string
		Cost  float64
	}
	var costs []qCost
	for _, sql := range queryList {
		p, err := db.Explain(sql.Text)
		if err != nil {
			return err
		}
		totCost += p.PlanCost()
		costs = append(costs, qCost{Alias: sql.Alias, Cost: p.PlanCost()})
	}
	sort.Slice(costs, func(i, j int) bool {
		return costs[i].Cost > costs[j].Cost
	})
	for _, c := range costs {
		fmt.Printf("%v %.2f %.2f\n", c.Alias, c.Cost/totCost*100, c.Cost)
	}
	return nil
}

func executeQueries(db optimizer.WhatIfOptimizer, queries utils.Set[utils.Query], savePath string) error {
	queryList := queries.ToList()
	sort.Slice(queryList, func(i, j int) bool {
		return queryList[i].Alias < queryList[j].Alias
	})

	os.MkdirAll(savePath, 0777)
	summaryContent := ""
	var totExecTime time.Duration
	for _, sql := range queryList {
		var execTimes []time.Duration
		var plans []utils.Plan
		for k := 0; k < 3; k++ {
			p, err := db.ExplainAnalyze(sql.Text)
			if err != nil {
				return err
			}
			plans = append(plans, p)
			execTimes = append(execTimes, p.ExecTime())
			fmt.Println(">> ", k, sql.Alias, p.ExecTime())
		}
		sort.Slice(execTimes, func(i, j int) bool {
			return execTimes[i] < execTimes[j]
		})
		avgTime := execTimes[1]
		totExecTime += avgTime

		content := fmt.Sprintf("Alias: %s\n", sql.Alias)
		content += fmt.Sprintf("AvgTime: %v\n", avgTime)
		content += fmt.Sprintf("ExecTimes: %v\n", execTimes)
		content += fmt.Sprintf("Query:\n %s\n\n", sql.Text)
		for _, p := range plans {
			content += fmt.Sprintf("%v\n", p.Format())
		}
		utils.SaveContentTo(fmt.Sprintf("%v/%v.txt", savePath, sql.Alias), content)

		summaryContent += fmt.Sprintf("%v %v\n", sql.Alias, avgTime)
		fmt.Println(sql.Alias, avgTime)
	}
	fmt.Println("TotalExecutionTime:", totExecTime)
	summaryContent += fmt.Sprintf("TotalExecutionTime: %v\n", totExecTime)
	return utils.SaveContentTo(fmt.Sprintf("%v/summary.txt", savePath), summaryContent)
}
