package cmd

import (
	"errors"
	"github.com/qw4990/index_advisor/advisor"
	"github.com/qw4990/index_advisor/optimizer"
	"github.com/qw4990/index_advisor/utils"
	"github.com/spf13/cobra"
)

type adviseOnlineCmdOpt struct {
	maxNumIndexes int
	maxIndexWidth int

	dsn      string
	output   string
	logLevel string

	querySchemas            []string
	queryExecTimeThreshold  int
	queryExecCountThreshold int
	queryPath               string
}

func NewAdviseOnlineCmd() *cobra.Command {
	var opt adviseOnlineCmdOpt
	cmd := &cobra.Command{
		Use:   "advise-online",
		Short: "advise some indexes for the specified workload",
		Long: `advise some indexes for the specified workload.
How it work:
1. connect to your online TiDB cluster through the DSN
2. read all queries from the 'STATEMENT_SUMMARY' system table
3. analyze those queries and generate a series of candidate indexes
4. evaluate those candidate indexes on your online TiDB cluster through a feature named 'hypothetical index' (or 'what-if index')
5. recommend you the best set of indexes based on the evaluation result
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.SetLogLevel(opt.logLevel)
			indexes, info, db, err := adviseOnlineMode(opt)
			if err != nil {
				return err
			}
			if indexes == nil {
				return nil
			}
			return outputAdviseResult(indexes, *info, db, opt.output)
		},
	}

	cmd.Flags().IntVar(&opt.maxNumIndexes, "max-num-indexes", 5, "max number of indexes to recommend, 1~20")
	cmd.Flags().IntVar(&opt.maxIndexWidth, "max-index-width", 3, "the max number of columns in recommended indexes")

	cmd.Flags().StringVar(&opt.dsn, "dsn", "root:@tcp(127.0.0.1:4000)/test", "dsn")
	cmd.Flags().StringVar(&opt.output, "output", "", "output directory to save the result")
	cmd.Flags().StringVar(&opt.logLevel, "log-level", "info", "log level, one of 'debug', 'info', 'warning', 'error'")

	cmd.Flags().StringSliceVar(&opt.querySchemas, "query-schemas", []string{}, "a list of schema(database), e.g. 'test1, test2', queries that are running under these schemas will be considered")
	cmd.Flags().IntVar(&opt.queryExecTimeThreshold, "query-exec-time-threshold", 0, "the threshold of query execution time(in milliseconds), e.g. '300', queries that are running longer than this threshold will be considered")
	cmd.Flags().IntVar(&opt.queryExecCountThreshold, "query-exec-count-threshold", 0, "the threshold of query execution count, e.g. '20', queries that are executed more than this threshold will be considered")
	cmd.Flags().StringVar(&opt.queryPath, "query-path", "", "the path that contains queries, e.g. 'queries.sql', if this variable is specified, the above variables like 'query-*' will be ignored")
	return cmd
}

func adviseOnlineMode(opt adviseOnlineCmdOpt) (utils.Set[utils.Index], *utils.WorkloadInfo, optimizer.WhatIfOptimizer, error) {
	db, err := optimizer.NewTiDBWhatIfOptimizer(opt.dsn)
	if err != nil {
		return nil, nil, nil, err
	}
	if reason := checkOnlineModeSupport(db); reason != "" {
		return nil, nil, nil, errors.New("online mode is not supported: " + reason)
	}

	info, err := prepareWorkloadOnlineMode(db, opt)
	if err != nil {
		return nil, nil, nil, err
	}
	if info.Queries.Size() == 0 {
		utils.Infof("no query is found")
		return nil, nil, nil, nil
	}

	result, err := advisor.IndexAdvise(db, *info, advisor.Parameter{
		MaxNumberIndexes: opt.maxNumIndexes,
		MaxIndexWidth:    opt.maxIndexWidth,
	})
	return result, info, db, err
}

func prepareWorkloadOnlineMode(db optimizer.WhatIfOptimizer, opt adviseOnlineCmdOpt) (*utils.WorkloadInfo, error) {
	var err error
	var queries utils.Set[utils.Query]
	if opt.queryPath == "" {
		queries, err = readQueriesFromStatementSummary(db, opt.querySchemas, opt.queryExecTimeThreshold, opt.queryExecCountThreshold)
		if err != nil {
			return nil, err
		}
		if queries.Size() == 0 {
			return nil, errors.New("no queries are found")
		}
	} else {
		_, dbName := utils.GetDBNameFromDSN(opt.dsn)
		if dbName == "" {
			return nil, errors.New("database name is not specified in DSN")
		}
		queries, err = utils.LoadQueries(dbName, opt.queryPath)
		if err != nil {
			return nil, err
		}
	}
	queries, err = filterSQLAccessingSystemTables(queries)
	if err != nil {
		return nil, err
	}
	tableNames, err := utils.CollectTableNamesFromQueries(queries)
	if err != nil {
		return nil, err
	}
	tables, err := getTableSchemas(db, tableNames)
	if err != nil {
		return nil, err
	}
	queries, err = filterSQLAccessingDroppedTable(queries, tables)
	if err != nil {
		return nil, err
	}
	return &utils.WorkloadInfo{
		Queries:      queries,
		TableSchemas: tables,
	}, nil
}
