package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pingcap/index_advisor/optimizer"
	"github.com/pingcap/index_advisor/utils"
)

// loadWorkloadIntoCluster loads the schema the TiDB cluster
func loadSchemaIntoCluster(db optimizer.WhatIfOptimizer, schemaFilePath string) (dbName string, err error) {
	if schemaFilePath == "" {
		return "", nil
	}
	utils.Infof("load schema info from %v into the TiDB instance", schemaFilePath)
	rawSQLs, err := utils.ParseStmtsFromFile(schemaFilePath)
	if err != nil {
		return "", err
	}
	if len(rawSQLs) == 0 {
		return "", nil
	}

	currentDB := "test" // the default DB `test`
	for _, stmt := range rawSQLs {
		switch utils.GetStmtType(stmt) {
		case utils.StmtUseDB:
			currentDB = utils.GetDBNameFromUseDBStmt(stmt)
		case utils.StmtCreateDB:
			dbName := utils.GetDBNameFromCreateDBStmt(stmt)
			exist, err := dbExists(dbName, db)
			if err != nil {
				return "", err
			}
			if exist {
				continue
			}
		case utils.StmtCreateTable:
			table, err := utils.ParseCreateTableStmt(currentDB, stmt)
			if err != nil {
				return "", err
			}
			utils.Infof("create table %s.%s", table.SchemaName, table.TableName)
		}
		if err := db.Execute(stmt); err != nil {
			return "", err
		}
	}
	return currentDB, nil
}

// loadStatsIntoCluster loads the stats into the TiDB cluster
func loadStatsIntoCluster(db optimizer.WhatIfOptimizer, statsDirPath string) error {
	if statsDirPath == "" {
		return nil
	}
	utils.Infof("load stats info from %v into the TiDB instance", statsDirPath)
	exist, _ := utils.FileExists(statsDirPath)
	if !exist {
		utils.Infof("no stats directory %s, skip loading stats", statsDirPath)
		return nil
	}
	utils.Infof("load stats from %s", statsDirPath)
	statsFiles, err := os.ReadDir(statsDirPath)
	if err != nil {
		return err
	}
	for _, statsFile := range statsFiles {
		statsPath := path.Join(statsDirPath, statsFile.Name())
		absStatsPath, err := filepath.Abs(statsPath)
		if err != nil {
			return err
		}
		tableName, err := getStatsFileTableName(absStatsPath)
		if err != nil {
			return err
		}

		utils.Infof("load stats for table %s from %s", tableName, statsPath)
		mysql.RegisterLocalFile(absStatsPath)
		loadStatsSQL := fmt.Sprintf("load stats '%s'", absStatsPath)
		if err := db.Execute(loadStatsSQL); err != nil {
			return err
		}
	}
	return nil
}

type statsJSON struct {
	DatabaseName string `json:"database_name"`
	TableName    string `json:"table_name"`
}

func getStatsFileTableName(statsFile string) (utils.TableName, error) {
	var stats statsJSON
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return utils.TableName{}, err
	}
	if err := json.Unmarshal(data, &stats); err != nil {
		return utils.TableName{}, err
	}
	return utils.TableName{stats.DatabaseName, stats.TableName}, nil
}

func tableExists(schemaName, tableName string, db optimizer.WhatIfOptimizer) (bool, error) {
	q := fmt.Sprintf("select count(*) from information_schema.TABLES where lower(table_schema) = '%s' and lower(table_name)='%s'",
		strings.ToLower(schemaName), strings.ToLower(tableName))
	r, err := db.Query(q)
	if err != nil {
		return false, err
	}
	r.Next()
	var count int
	if err := r.Scan(&count); err != nil {
		return false, err
	}
	if err := r.Close(); err != nil {
		return false, err
	}
	return count > 0, nil
}

func dbExists(schemaName string, db optimizer.WhatIfOptimizer) (bool, error) {
	q := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.SCHEMATA where lower(SCHEMA_NAME) = '%s'", strings.ToLower(schemaName))
	r, err := db.Query(q)
	if err != nil {
		return false, err
	}
	r.Next()
	var count int
	if err := r.Scan(&count); err != nil {
		return false, err
	}
	if err := r.Close(); err != nil {
		return false, err
	}
	return count > 0, nil
}

// supportHypoIndex tests whether this TiDB version supports hypothetical indexes.
func supportHypoIndex(db optimizer.WhatIfOptimizer) bool {
	err := db.Execute(`drop hypo index hypo_index_test_name on test`)
	if strings.Contains(err.Error(), "You have an error in your SQL syntax") {
		return false
	}
	// if it's not a syntax error, we assume it supports Hypo Indexes.
	return true
}

func redactLogEnabled(db optimizer.WhatIfOptimizer) bool {
	rows, err := db.Query(`select @@global.tidb_redact_log`)
	if err != nil {
		return false
	}
	defer rows.Close()
	if !rows.Next() {
		return false
	}
	var redactLog string
	if err := rows.Scan(&redactLog); err != nil {
		return false
	}
	redactLog = strings.ToLower(redactLog)
	if redactLog == "1" || redactLog == "on" {
		return true
	}
	return false
}

func readQueriesFromStatementSummary(db optimizer.WhatIfOptimizer, querySchemas []string,
	queryExecTimeThreshold, queryExecCountThreshold int) (utils.Set[utils.Query], error) {
	var condition []string
	condition = append(condition, "stmt_type='Select'")
	if len(querySchemas) > 0 {
		condition = append(condition, fmt.Sprintf("SCHEMA_NAME in ('%s')", strings.Join(querySchemas, "', '")))
	}
	if queryExecTimeThreshold > 0 {
		condition = append(condition, fmt.Sprintf("AVG_LATENCY >= %v", queryExecTimeThreshold*1000))
	}
	if queryExecCountThreshold > 0 {
		condition = append(condition, fmt.Sprintf("EXEC_COUNT >= %v", queryExecCountThreshold))
	}
	// TODO: consider Execute statements

	s := utils.NewSet[utils.Query]()
	for _, table := range []string{
		`information_schema.statements_summary`,
		`information_schema.statements_summary_history`,
	} {
		q := fmt.Sprintf(`select SCHEMA_NAME, DIGEST, QUERY_SAMPLE_TEXT, EXEC_COUNT, AVG_LATENCY from %v where %v`,
			table, strings.Join(condition, " AND "))
		rows, err := db.Query(q)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var schemaName, digest, text, execCountStr, avgLatStr sql.NullString
			if err := rows.Scan(&schemaName, &digest, &text, &execCountStr, &avgLatStr); err != nil {
				return nil, err
			}
			execCount, err := strconv.Atoi(execCountStr.String)
			if err != nil {
				return nil, err
			}
			if _, err := utils.ParseOneSQL(text.String); err != nil {
				// some queries may be truncated, we skip them.
				continue
			}

			// TODO: what if this query's database has been dropped?
			// TODO: skip this query if it has '?' when redact log is enabled.
			s.Add(utils.Query{
				Alias:      digest.String,
				SchemaName: schemaName.String, // can be empty (null)
				Text:       text.String,
				Frequency:  execCount,
			})
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func readTableSchemas(db optimizer.WhatIfOptimizer, schemas []string) (utils.Set[utils.TableSchema], error) {
	s := utils.NewSet[utils.TableSchema]()
	for _, schemaName := range schemas {
		tableNames, err := readTableNames(db, schemaName)
		if err != nil {
			return nil, err
		}
		for _, tableName := range tableNames {
			q := fmt.Sprintf(`show create table %s.%s`, schemaName, tableName)
			rows, err := db.Query(q)
			if err != nil {
				return nil, err
			}
			for rows.Next() {
				var name, createTableStmt string
				if err := rows.Scan(&name, &createTableStmt); err != nil {
					return nil, err
				}
				tableSchema, err := utils.ParseCreateTableStmt(schemaName, createTableStmt)
				if err != nil {
					return nil, err
				}
				s.Add(tableSchema)
			}
			rows.Close()
		}
	}
	return s, nil
}

func readTableNames(db optimizer.WhatIfOptimizer, schemaName string) ([]string, error) {
	if err := db.Execute(fmt.Sprintf(`use %s`, schemaName)); err != nil {
		return nil, err
	}
	q := `show tables`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}
	return tableNames, nil
}

func filterSQLAccessingSystemTables(sqls utils.Set[utils.Query]) (utils.Set[utils.Query], error) {
	s := utils.NewSet[utils.Query]()
	for _, sql := range sqls.ToList() {
		accessSystemTable := false
		tables, err := utils.CollectTableNamesFromSQL(sql.SchemaName, sql.Text)
		if err != nil {
			return nil, err
		}
		if tables.Size() == 0 {
			// `select @@some_var` or `select some_func()`
			continue
		}
		for _, t := range tables.ToList() {
			if utils.IsTiDBSystemTableName(t) {
				accessSystemTable = true
				break
			}
		}
		if !accessSystemTable {
			s.Add(sql)
		}
	}
	return s, nil
}

func filterSQLAccessingDroppedTable(sqls utils.Set[utils.Query], tables utils.Set[utils.TableSchema]) (utils.Set[utils.Query], error) {
	tableNames := utils.NewSet[utils.TableName]()
	for _, t := range tables.ToList() {
		tableNames.Add(utils.TableName{
			SchemaName: t.SchemaName,
			TableName:  t.TableName,
		})
	}

	s := utils.NewSet[utils.Query]()
	for _, sql := range sqls.ToList() {
		tables, err := utils.CollectTableNamesFromSQL(sql.SchemaName, sql.Text)
		if err != nil {
			return nil, err
		}
		noTableFlag := false
		for _, t := range tables.ToList() {
			if !tableNames.Contains(utils.TableName{
				SchemaName: t.SchemaName,
				TableName:  t.TableName,
			}) {
				noTableFlag = true
				break
			}
		}
		if noTableFlag {
			continue
		}
		s.Add(sql)
	}
	return s, nil
}
