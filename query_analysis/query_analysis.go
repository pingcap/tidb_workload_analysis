package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	sqlite "github.com/mattn/go-sqlite3"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

func fatal(errMsg string) {
	fmt.Println(errMsg)
	os.Exit(0)
}

func exec(c *sql.Conn, sql string) {
	_, err := c.ExecContext(context.Background(), sql)
	if err != nil {
		fatal(fmt.Sprintf("execute %s failed: %s", sql, err.Error()))
	}
}

func query(c *sql.Conn, sql string, drainer func(*sql.Rows)) {
	r, err := c.QueryContext(context.Background(), sql)
	if err != nil {
		fatal(fmt.Sprintf("query %s failed: %s", sql, err.Error()))
	}
	drainer(r)
	if err := r.Close(); err != nil {
		fatal(fmt.Sprintf("close query %s failed: %s", sql, err.Error()))
	}
}

func reportByFrequency(c *sql.Conn) {
	q := `select query_type, sum(frequency) from unique_queries group by query_type order by sum(frequency) desc`
	query(c, q, func(r *sql.Rows) {
		fmt.Printf("%-35s%s\n", "Query Type", "Frequency")
		var queryType string
		var frequency int
		for r.Next() {
			if err := r.Scan(&queryType, &frequency); err != nil {
				fatal("scan failed: " + err.Error())
			}
			fmt.Printf("%-35s%d\n", queryType, frequency)
		}
		fmt.Println()
	})
}

func reportByQueryResource(c *sql.Conn, timeOrMem string) {
	col1 := "Query Type"
	var col2, q string
	if timeOrMem == "time" {
		col2 = "Total Query Time in Seconds"
		q = "select query_type, sum(total_query_time) from unique_queries group by query_type order by sum(total_query_time) desc"
	} else if timeOrMem == "memory" {
		col2 = "Total MB Memory"
		q = "select query_type, sum(total_mem) from unique_queries group by query_type order by sum(total_mem) desc"
	} else {
		fatal("invalid timeOrMem")
	}

	query(c, q, func(r *sql.Rows) {
		fmt.Printf("%-35s%s\n", col1, col2)
		var queryType string
		var total float64
		for r.Next() {
			if err := r.Scan(&queryType, &total); err != nil {
				fatal("scan failed: " + err.Error())
			}
			fmt.Printf("%-35s%.2f\n", queryType, total)
		}
		fmt.Println()
	})
}

func readVsWriteReport(c *sql.Conn) {
	q := `select query_markers, frequency, total_query_time, total_mem from unique_queries`
	query(c, q, func(r *sql.Rows) {
		var readFreq, writeFreq int
		var readTime, writeTime float64
		var readMem, writeMem float64

		for r.Next() {
			var markers string
			var frequency int
			var totalQueryTime, totalMem float64
			err := r.Scan(&markers, &frequency, &totalQueryTime, &totalMem)
			if err != nil {
				fatal("scan failed: " + err.Error())
			}
			if strings.Contains(strings.ToLower(markers), "select") { // read
				readFreq += frequency
				readTime += totalQueryTime
				readMem += totalMem
			} else { // write
				writeFreq += frequency
				writeTime += totalQueryTime
				writeMem += totalMem
			}
		}
		readTime = math.Ceil(readTime) * 100 / 100
		writeTime = math.Ceil(writeTime) * 100 / 100
		readMem = math.Ceil(readMem/1024/1024) * 100 / 100
		writeMem = math.Ceil(writeMem/1024/1024) * 100 / 100
		fmt.Printf("%-35s%-20s%-20s%-20s\n", "Query Type", "Frequency", "Total Time", "Total MB Memory")
		fmt.Printf("%-35s%-20d%-20s%-20s\n", "Read", readFreq, fmt.Sprintf("%.2f", readTime), fmt.Sprintf("%.2f", readMem))
		fmt.Printf("%-35s%-20d%-20s%-20s\n", "Write", writeFreq, fmt.Sprintf("%.2f", writeTime), fmt.Sprintf("%.2f", writeMem))
		fmt.Println()
	})
}

func insertSelectVsInsertValuesReport(c *sql.Conn) {
	q := `select query_type, frequency, total_query_time, total_mem  
            from unique_queries
            where query_markers like '%Insert%'`

	query(c, q, func(r *sql.Rows) {
		var insertSelectFre, insertValuesFre int
		var insertSelectTime, insertValuesTime float64
		var insertSelectMem, insertValuesMem float64

		for r.Next() {
			var queryType string
			var frequency int
			var totalQueryTime, totalMem float64
			err := r.Scan(&queryType, &frequency, &totalQueryTime, &totalMem)
			if err != nil {
				fatal("scan failed: " + err.Error())
			}
			if strings.Contains(strings.ToLower(queryType), "insert_values") {
				insertValuesFre += frequency
				insertValuesTime += totalQueryTime
				insertValuesMem += totalMem
			} else {
				insertSelectFre += frequency
				insertSelectTime += totalQueryTime
				insertSelectMem += totalMem
			}
		}
		insertSelectTime = math.Ceil(insertSelectTime) * 100 / 100
		insertValuesTime = math.Ceil(insertValuesTime) * 100 / 100
		insertSelectMem = math.Ceil(insertSelectMem/1024/1024) * 100 / 100
		insertValuesMem = math.Ceil(insertValuesMem/1024/1024) * 100 / 100

		fmt.Printf("%-35s%-20s%-20s%-20s\n", "Query Type", "Frequency", "Total Time", "Total MB Memory")
		fmt.Printf("%-35s%-20d%-20s%-20s\n", "Insert Select", insertSelectFre, fmt.Sprintf("%.2f", insertSelectTime), fmt.Sprintf("%.2f", insertSelectMem))
		fmt.Printf("%-35s%-20d%-20s%-20s\n", "Insert Values", insertValuesFre, fmt.Sprintf("%.2f", insertValuesTime), fmt.Sprintf("%.2f", insertValuesMem))
		fmt.Println()
	})
}

func analyzeQueries(c *sql.Conn) {
	exec(c, "drop table if exists unique_queries")
	exec(c, `create table unique_queries(
                digest text,
                sql_statement text,
                frequency decimal,
                total_query_time decimal,
                max_query_time decimal,
                min_query_time decimal,
                total_mem decimal,
                max_mem decimal,
                min_mem decimal,
                query_markers text,
                number_of_joins tinyint,
                query_type text)`)
	exec(c, `INSERT INTO unique_queries(
                digest, sql_statement, frequency,
                total_query_time, max_query_time, min_query_time,
                total_mem, max_mem, min_mem,
                query_markers, number_of_joins, query_type
            )
            SELECT digest, max(query_sample_text), sum(exec_count) as frequency,
                sum(CAST(sum_latency as decimal)) as total_query_time,
                max(CAST(max_latency as decimal)) as max_query_time,
                min(CAST(min_latency as decimal)) as min_query_time,
                sum(CAST((avg_mem*exec_count) as decimal)) as total_mem,
                max(CAST(max_mem as decimal)) as max_mem,
                min(CAST(avg_mem as decimal)) as min_mem,
                analyzeOneQuery(query_sample_text),
                numberOfJoins(analyzeOneQuery(query_sample_text)),
                queryType(analyzeOneQuery(query_sample_text))
            FROM statements_summary
            GROUP BY digest`)

	reportByFrequency(c)
	reportByQueryResource(c, "time")
	reportByQueryResource(c, "memory")
	readVsWriteReport(c)
	insertSelectVsInsertValuesReport(c)
}

func analyzeOneQuery(sql string) string {
	// few hacks to clean the sql
	if strings.HasPrefix(sql, "-- Metabase") {
		sql = sql[len("-- Metabase"):]
	}
	return getSQLTypeList(sql)
}

func numberOfJoins(typeListStr string) int {
	return strings.Count(typeListStr, "InnerJoin") +
		strings.Count(typeListStr, "LeftJoin") +
		strings.Count(typeListStr, "RightJoin")
}

func queryType(typeListStr string) string {
	k := typeListStr
	c := func(k, subKey string) bool {
		return strings.Contains(strings.ToLower(k), strings.ToLower(subKey))
	}
	if c(k, "insert") {
		return "INSERT_VALUES"
	} else if c(k, "delete") && !c(k, "select") {
		return "DELETE"
	} else if c(k, "update") && !c(k, "select") {
		return "UPDATE"
	} else if c(k, "analyze") {
		return "ANALYZE"
	} else if c(k, "none") {
		return "UNKNOWN"
	} else if c(k, "explain") {
		return "EXPLAIN"
	} else if c(k, "system") {
		return "SYSTEM"
	} else {
		numAgg := numberOfAggregate(k)
		numGroupBy := numberOfGroupBy(k)
		numJoins := numberOfJoins(k)
		if numAgg == 0 && numGroupBy == 0 && numJoins == 0 {
			return "SCAN"
		} else if numAgg > 0 && numGroupBy == 0 && numJoins == 0 {
			return "AGGREGATE_SCAN_NO_GROUPBY"
		} else if numAgg > 0 && numGroupBy == 0 && numJoins > 0 {
			return "AGGREGATE_JOIN_NO_GROUPBY"
		} else if numGroupBy > 0 && numJoins == 0 {
			return "AGGREGATE_SCAN_GROUPBY"
		} else if numGroupBy > 0 && numJoins > 0 {
			return "AGGREGATE_JOIN_GROUPBY"
		} else if numAgg == 0 && numGroupBy == 0 && numJoins > 0 {
			return "JOIN_NO_AGGREGATE"
		} else {
			return "FIX-IT"
		}
	}
}

func numberOfAggregate(typeListStr string) int {
	return strings.Count(typeListStr, "Aggregate")
}

func numberOfGroupBy(typeListStr string) int {
	return strings.Count(typeListStr, "GroupBy") +
		strings.Count(typeListStr, "Distinct")
}

func main() {
	sql.Register("sqlite3_custom", &sqlite.SQLiteDriver{
		ConnectHook: func(conn *sqlite.SQLiteConn) error {
			if err := conn.RegisterFunc("analyzeOneQuery", analyzeOneQuery, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("numberOfJoins", numberOfJoins, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("queryType", queryType, true); err != nil {
				return err
			}
			return nil
		},
	})

	var testDatabase string
	flag.StringVar(&testDatabase, "test_database", "", "The database to test")
	flag.Parse()
	if testDatabase == "" {
		fatal("no test_database")
	}

	db, err := sql.Open("sqlite3_custom", testDatabase)
	if err != nil {
		fatal("Open database failed: " + err.Error())
	}
	c, err := db.Conn(context.Background())
	if err != nil {
		fatal("Open connection failed: " + err.Error())
	}
	analyzeQueries(c)
}

// column analysis visitor
type columnAnalysis struct {
	columnCount uint64
}

// analyzeColumns goes through an expression and returns a count of column references.
// The second output is a boolean indicating success or failure of the visitor.
func (v *columnAnalysis) analyzeColumns(node ast.Node) (uint64, bool) {
	switch node.(type) {
	case *ast.ColumnName, *ast.ColumnDef:
		return 1, true
	// Temporary fix since visitor is not working for extract. TODO: fix visitor.
	case *ast.TimeUnitExpr:
		return 1, true
	}
	return 0, true
}

func (v *columnAnalysis) Enter(in ast.Node) (ast.Node, bool) {
	if colCount, ok := v.analyzeColumns(in); ok {
		v.columnCount = v.columnCount + colCount
	}
	return in, false
}

func (v *columnAnalysis) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func columnVisitor(rootNode ast.Node) uint64 {
	v := &columnAnalysis{}
	rootNode.Accept(v)
	return v.columnCount
}

// end of column analysis visitor

// Visitor to collect node types from a SQL query.
type typeAnalysis struct {
	typesList []string
}

// Main function for the typeAnalysis visitor with an ast node as input.
// The second output parameter is a boolean
// indicating failure/success. The first output parameter is a list of types collected
// from the SQL and it is empty if there is a failure.
func (v *typeAnalysis) analyzeTypes(node ast.Node) ([]string, bool) {
	switch nodeType := node.(type) {
	case *ast.ExplainStmt, *ast.ExplainForStmt:
		return []string{"Explain"}, true
	case *ast.ShowStmt, *ast.SetStmt, *ast.UseStmt,
		*ast.BeginStmt, *ast.CommitStmt,
		*ast.RollbackStmt, *ast.CreateUserStmt, *ast.SetPwdStmt:
		return []string{"System"}, true
	case *ast.AnalyzeTableStmt:
		return []string{"Analyze"}, true
	case *ast.DeleteStmt:
		return []string{"Delete"}, true
	case *ast.UpdateStmt:
		return []string{"Update"}, true
	case *ast.InsertStmt:
		return []string{"Insert"}, true
	case *ast.SubqueryExpr, *ast.ExistsSubqueryExpr:
		return []string{"InnerJoin"}, true
	case *ast.PatternInExpr:
		if nodeType.Sel == nil {
			return []string{"Filter"}, true
		}
	case *ast.AggregateFuncExpr:
		if nodeType.Distinct {
			return []string{"Aggregate", "Distinct"}, true
		}
		return []string{"Aggregate"}, true
	case ast.ExprNode:
		columnCount := columnVisitor(node)
		_, valueExpr := node.(ast.ValueExpr)
		if !valueExpr && (columnCount == 0) {
			return []string{"Constant"}, true
		}
	case *ast.OrderByClause:
		return []string{"OrderBy"}, true
	case *ast.GroupByClause:
		return []string{"GroupBy"}, true
	case *ast.SelectStmt:
		if nodeType.AfterSetOperator != nil && (*nodeType.AfterSetOperator == ast.Union || *nodeType.AfterSetOperator == ast.UnionAll) {
			return []string{"Select", "Union"}, true
		}
		markers := []string{"Select"}
		if nodeType.Where != nil {
			markers = append(markers, "Filter")
		}
		if nodeType.Distinct {
			markers = append(markers, "Distinct")
		}
		if len(markers) > 0 {
			return markers, true
		}
	case *ast.Join:
		if nodeType.Right != nil {
			switch nodeType.Tp {
			case ast.LeftJoin:
				return []string{"LeftJoin"}, true
			case ast.RightJoin:
				return []string{"RightJoin"}, true
			default:
				return []string{"InnerJoin"}, true
			}
		}
	case *ast.HavingClause:
		return []string{"Having"}, true
	}
	return []string{"Other"}, false
}

func (v *typeAnalysis) Enter(in ast.Node) (ast.Node, bool) {
	if func_type, ok := v.analyzeTypes(in); ok {
		v.typesList = append(v.typesList, func_type...)
	}
	return in, false
}

func (v *typeAnalysis) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

// wrap typeAnalysis
func typeVisitor(rootNode ast.StmtNode) []string {
	v := &typeAnalysis{}
	rootNode.Accept(v)
	return v.typesList
}

// wrapper around TiDB parser which takes a SQL and returns
// a parse statement and error code.
func parse(sql string) (ast.StmtNode, error) {
	p := parser.New()
	stmtNodes, _, err := p.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}
	if len(stmtNodes) == 0 {
		return nil, nil
	} else {
		return stmtNodes[0], nil
	}
}

// Main function that takes a SQL query and returns a list of node types/markers.
func getSQLTypeList(sql string) (typeListStr string) {
	defer func() {
		if typeListStr == "" {
			sql := strings.ToLower(sql)
			if matched, err := regexp.Match(".*insert.*values.*", []byte(sql)); err == nil && matched {
				typeListStr = "Insert"
			} else if matched, err := regexp.Match("^analyze.*", []byte(sql)); err == nil && matched {
				typeListStr = "Analyze"
			} else {
				typeListStr = "None"
			}
		}
	}()

	astNode, err := parse(sql)
	if err != nil || astNode == nil {
		return ""
	}

	typeList := typeVisitor(astNode)
	var m = make(map[string]bool)
	var a []string
	for _, value := range typeList {
		if !m[value] {
			a = append(a, value)
			m[value] = true
		}
	}
	return strings.Join(typeList, ",")
}
