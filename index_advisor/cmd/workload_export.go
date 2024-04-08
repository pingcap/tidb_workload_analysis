package cmd

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/pingcap/index_advisor/optimizer"
	"github.com/pingcap/index_advisor/utils"
	"github.com/spf13/cobra"
)

type workloadExportCmdOpt struct {
	dsn        string
	statusAddr string
	output     string
	logLevel   string
}

func NewWorkloadExportCmd() *cobra.Command {
	var opt workloadExportCmdOpt
	cmd := &cobra.Command{
		Use:   "workload-export",
		Short: "export workload information (queries, table schema, table statistics) from your TiDB cluster, use `index_advisor workload-export --help` to see more details`",
		Long: `export workload information (queries, table schema, table statistics) from your TiDB cluster.
How it work:
1. connect to your TiDB cluster through the DSN
2. read all queries from the 'STATEMENT_SUMMARY' system table
3. read all table schema from the 'INFORMATION_SCHEMA' database
4. read all statistics from the 'mysql.stats_xxx' system tables
5. store all data into the specified output directory
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.SetLogLevel(opt.logLevel)
			utils.Infof("[workload-export] start exporting workload information from TiDB cluster %v to %v", opt.dsn, opt.output)
			err := exportWorkload(opt)
			if err == nil {
				utils.Infof("[workload-export] export workload information successfully into %s", opt.output)
			} else {
				utils.Infof("[workload-export] export workload information failed: %v", err)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&opt.dsn, "dsn", "root:@tcp(127.0.0.1:4000)/test", "dsn")
	cmd.Flags().StringVar(&opt.statusAddr, "status_address", "http://127.0.0.1:10080", "status address used to download table statistics")
	cmd.Flags().StringVar(&opt.output, "output", "", "output directory to save the result")
	cmd.Flags().StringVar(&opt.logLevel, "log-level", "info", "log level, one of 'debug', 'info', 'warning', 'error'")
	return cmd
}

func exportWorkload(opt workloadExportCmdOpt) error {
	utils.Infof("[workload-export] prepare dir %v", opt.output)
	if err := utils.PrepareDir(opt.output); err != nil {
		return err
	}
	utils.Infof("[workload-export] connect to %v", opt.dsn)
	db, err := optimizer.NewTiDBWhatIfOptimizer(opt.dsn)
	if err != nil {
		return err
	}
	queries, err := readQueriesFromStatementSummary(db, nil, 0, 0)
	if err != nil {
		return err
	}
	queries, err = filterSQLAccessingSystemTables(queries)
	if err != nil {
		return err
	}
	tableNames, err := utils.CollectTableNamesFromQueries(queries)
	if err != nil {
		return err
	}
	tables, err := getTableSchemas(db, tableNames)
	if err != nil {
		return err
	}
	queries, err = filterSQLAccessingDroppedTable(queries, tables)
	if err != nil {
		return err
	}
	utils.Infof("[workload-export] read %v queries", queries.Size())
	if err := saveQueries(opt, queries); err != nil {
		return err
	}
	if err := saveTableSchemas(opt, tables); err != nil {
		return err
	}

	utils.Infof("[workload-export] start dumping table statistics for %v tables", tables.Size())
	statsDir := path.Join(opt.output, "stats")
	utils.Infof("[workload-export] prepare stats dir %v", statsDir)
	if err := utils.PrepareDir(statsDir); err != nil {
		return err
	}
	for _, t := range tables.ToList() {
		stats, err := fetchTableStats(opt, t)
		if err != nil {
			return err
		}
		fpath := path.Join(statsDir, fmt.Sprintf("%s_%s.json", t.SchemaName, t.TableName))
		if err := utils.SaveContentTo(fpath, string(stats)); err != nil {
			return err
		}
		utils.Infof("[workload-export] save table statistics for %v to %v", t.Key(), fpath)
	}
	return nil
}

func fetchTableStats(opt workloadExportCmdOpt, table utils.TableSchema) ([]byte, error) {
	// http://${tidb-server-ip}:${tidb-server-status-port}/stats/dump/${db_name}/${table_name}
	url := fmt.Sprintf("%s/stats/dump/%s/%s", opt.statusAddr, table.SchemaName, table.TableName)
	stats, err := utils.ReadURL(url)
	if err != nil {
		utils.Infof("[workload-export] fail to dump statistics for %v from %v, err: %v", table.Key(), url, err)
		return nil, err
	}
	utils.Infof("[workload-export] succeed to dump statistics for %v from %v", table.Key(), url)
	return stats, err
}

func saveQueries(opt workloadExportCmdOpt, queries utils.Set[utils.Query]) error {
	var buf bytes.Buffer
	for _, q := range queries.ToList() {
		if q.SchemaName != "" {
			buf.WriteString(fmt.Sprintf("use %s;\n", q.SchemaName))
		}
		text := strings.TrimSpace(q.Text)
		buf.WriteString(text)
		if !strings.HasSuffix(text, ";") {
			buf.WriteString(";")
		}
		buf.WriteString("\n\n")
	}
	fpath := path.Join(opt.output, "queries.sql")
	utils.Infof("[workload-export] save queries to %v", fpath)
	return utils.SaveContentTo(fpath, buf.String())
}

func saveTableSchemas(opt workloadExportCmdOpt, tables utils.Set[utils.TableSchema]) error {
	var buf bytes.Buffer
	for _, t := range tables.ToList() {
		buf.WriteString(fmt.Sprintf("create database if not exists %s;\n", t.SchemaName))
		buf.WriteString(fmt.Sprintf("use %s;\n", t.SchemaName))
		text := strings.TrimSpace(t.CreateStmtText)
		buf.WriteString(text)
		if !strings.HasSuffix(text, ";") {
			buf.WriteString(";")
		}
		buf.WriteString("\n\n")
	}
	fpath := path.Join(opt.output, "schema.sql")
	utils.Infof("[workload-export] save table schema into %s", fpath)
	return utils.SaveContentTo(fpath, buf.String())
}
