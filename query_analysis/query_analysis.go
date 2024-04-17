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
		fmt.Printf("%-35s%-15s%-15s%-15s\n", "Query Type", "Frequency", "Total Time", "Total MB Memory")
		fmt.Printf("%-35s%-15d%-15s%-15s\n", "Read", readFreq, fmt.Sprintf("%.2f", readTime), fmt.Sprintf("%.2f", readMem))
		fmt.Printf("%-35s%-15d%-15s%-15s\n", "Write", writeFreq, fmt.Sprintf("%.2f", writeTime), fmt.Sprintf("%.2f", writeMem))
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

		fmt.Printf("%-35s%-15s%-15s%-15s\n", "Query Type", "Frequency", "Total Time", "Total MB Memory")
		fmt.Printf("%-35s%-15d%-15s%-15s\n", "Insert Select", insertSelectFre, fmt.Sprintf("%.2f", insertSelectTime), fmt.Sprintf("%.2f", insertSelectMem))
		fmt.Printf("%-35s%-15d%-15s%-15s\n", "Insert Values", insertValuesFre, fmt.Sprintf("%.2f", insertValuesTime), fmt.Sprintf("%.2f", insertValuesMem))
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
	keyString := analyze_driver(sql, true)
	if keyString == "" {
		sql := strings.ToLower(sql)
		if matched, err := regexp.Match(".*insert.*values.*", []byte(sql)); err == nil && matched {
			keyString = "[Insert]"
		} else if matched, err := regexp.Match("^analyze.*", []byte(sql)); err == nil && matched {
			keyString = "[Analyze]"
		} else {
			keyString = "[None]"
		}
	}
	return keyString
}

func numberOfJoins(keyString string) int {
	return strings.Count(keyString, "InnerJoin") +
		strings.Count(keyString, "LeftJoin") +
		strings.Count(keyString, "RightJoin")
}

func queryType(k string) string {
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

func numberOfAggregate(keyString string) int {
	return strings.Count(keyString, "Aggregate")
}

func numberOfGroupBy(keyString string) int {
	return strings.Count(keyString, "GroupBy") +
		strings.Count(keyString, "Distinct")
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
