package utils

import (
	"fmt"
	"github.com/pingcap/parser/ast"
	"strings"
)

// FilterQueries filters Queries by their alias.
func FilterQueries(sqls Set[Query], whiteList, blackList []string) Set[Query] {
	whiteMap := make(map[string]bool)
	for _, a := range whiteList {
		if strings.TrimSpace(a) != "" {
			whiteMap[strings.TrimSpace(a)] = true
		}
	}
	blackMap := make(map[string]bool)
	for _, a := range blackList {
		if strings.TrimSpace(a) != "" {
			blackMap[strings.TrimSpace(a)] = true
		}
	}

	filtered := NewSet[Query]()
	for _, sql := range sqls.ToList() {
		if (len(whiteMap) > 0 && !whiteMap[sql.Alias]) ||
			(len(blackMap) > 0 && blackMap[sql.Alias]) {
			continue
		}
		filtered.Add(sql)
	}
	return filtered
}

// CreateWorkloadFromRawStmt creates a WorkloadInfo from some raw Queries.
func CreateWorkloadFromRawStmt(schemaName string, createTableStmts, rawSQLs []string) (WorkloadInfo, error) {
	sqls := NewSet[Query]()
	for _, rawSQL := range rawSQLs {
		sqls.Add(Query{
			SchemaName: schemaName,
			Text:       rawSQL,
			Frequency:  1,
		})
	}
	tableSchemas := NewSet[TableSchema]()
	for _, createStmt := range createTableStmts {
		tableSchema, err := ParseCreateTableStmt(schemaName, createStmt)
		if err != nil {
			return WorkloadInfo{}, err
		}
		tableSchemas.Add(tableSchema)
	}
	return WorkloadInfo{
		Queries:      sqls,
		TableSchemas: tableSchemas,
	}, nil
}

// LoadQueries loads queries from the given path.
func LoadQueries(schemaName, queryPath string) (Set[Query], error) {
	queries := NewSet[Query]()
	if exist, isDir := FileExists(queryPath); exist && isDir {
		rawSQLs, names, err := ParseStmtsFromDir(queryPath)
		if err != nil {
			return nil, err
		}
		for i, rawSQL := range rawSQLs {
			queries.Add(Query{
				Alias:      strings.Split(names[i], ".")[0], // q1.sql, 2a.sql, etc.
				SchemaName: schemaName,                      // Notice: for simplification, assume all Queries are under the same schema here.
				Text:       rawSQL,
				Frequency:  1,
			})
		}
		Infof("load %d queries from dir %s", len(rawSQLs), queryPath)
	} else if exist, isDir := FileExists(queryPath); exist || !isDir {
		rawSQLs, err := ParseStmtsFromFile(queryPath)
		if err != nil {
			return nil, err
		}
		for i, rawSQL := range rawSQLs {
			stmtType := GetStmtType(rawSQL)
			if stmtType == StmtUseDB {
				schemaName = GetDBNameFromUseDBStmt(rawSQL)
			}
			if stmtType != StmtSelect {
				continue
			}

			queries.Add(Query{
				Alias:      fmt.Sprintf("q%v", i+1),
				SchemaName: schemaName, // Notice: for simplification, assume all Queries are under the same schema here.
				Text:       rawSQL,
				Frequency:  1,
			})
		}
		Infof("load %d queries from %s", len(rawSQLs), queryPath)
	} else {
		return nil, fmt.Errorf("can not find queries directory or queries.sql file under %s", queryPath)
	}
	return queries, nil
}

// ParseCreateTableStmt parses a create table statement and returns a TableSchema.
func ParseCreateTableStmt(schemaName, createTableStmt string) (TableSchema, error) {
	stmt, err := ParseOneSQL(createTableStmt)
	if err != nil {
		return TableSchema{}, err
	}
	createTable := stmt.(*ast.CreateTableStmt)
	t := TableSchema{
		SchemaName:     schemaName,
		TableName:      createTable.Table.Name.L,
		CreateStmtText: createTableStmt,
	}
	for _, colDef := range createTable.Cols {
		t.Columns = append(t.Columns, Column{
			SchemaName: schemaName,
			TableName:  createTable.Table.Name.L,
			ColumnName: colDef.Name.Name.L,
			ColumnType: colDef.Tp.Clone(),
		})
	}
	// TODO: parse indexes
	return t, nil
}

// ParseCreateIndexStmt parses a create index statement and returns an Index.
func ParseCreateIndexStmt(createIndexStmt string) (Index, error) {
	stmt, err := ParseOneSQL(createIndexStmt)
	if err != nil {
		return Index{}, err
	}
	createIndex := stmt.(*ast.CreateIndexStmt)
	schemaName, tableName := createIndex.Table.Schema.O, createIndex.Table.Name.O
	if schemaName == "" {
		return Index{}, fmt.Errorf("schema name is empty")
	}
	index := Index{
		SchemaName: schemaName,
		TableName:  tableName,
		IndexName:  createIndex.IndexName,
	}
	for _, col := range createIndex.IndexPartSpecifications {
		index.Columns = append(index.Columns, Column{
			SchemaName: schemaName,
			TableName:  tableName,
			ColumnName: col.Column.Name.O,
		})
	}
	return index, nil
}
