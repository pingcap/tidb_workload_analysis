package cmd

import (
	"fmt"
	"os"
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
				utils.Infof("%d queries left after filtering", queries.Size())
			}

			db, err := optimizer.NewTiDBWhatIfOptimizer(opt.dsn)
			if err != nil {
				return err
			}

			sqls := queries.ToList()
			sort.Slice(sqls, func(i, j int) bool {
				return sqls[i].Alias < sqls[j].Alias
			})

			if opt.analyze {
				return explainQueries(db, queries)
			} else {
				return executeQueries(db, queries, opt.output)
			}
		},
	}

	cmd.Flags().StringVar(&opt.dsn, "dsn", "root:@tcp(127.0.0.1:4000)/test", "dsn")
	cmd.Flags().BoolVar(&opt.analyze, "analyze", true, "whether to use `explain analyze`")
	cmd.Flags().StringVar(&opt.queryPath, "query-path", "", "")
	cmd.Flags().StringVar(&opt.qWhiteList, "query-white-list", "", "queries to consider, e.g. 'q1,q2,q6'")
	cmd.Flags().StringVar(&opt.qBlackList, "query-black-list", "", "queries to ignore, e.g. 'q5,q12'")
	cmd.Flags().StringVar(&opt.output, "output", "", "output directory to save the result")
	return cmd
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
