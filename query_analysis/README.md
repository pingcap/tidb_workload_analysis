# Query Analysis (Beta)

## Objective and scope
This tool examines the structure of workload queries. This analysis serves both analysts and engineers, providing valuable insights. Furthermore, it can facilitate the simulation of user query workloads.

## How to run the tool
```
go run query_analysis.go --test_database <querylog_database>
```

## Prerequisites 
- Golang should be installed on the system
- Need input database in sqlite format that has a table called slowlog. slowlog table
  can be constructed from slowlog using O11 tool. 
- Need to install sqlite3 since the tool uses it to access the log data. 

## Sample test
We have some sample data in `data_normalizer/testdata`, you can use that to test the tool.

```
# example 1
go run query_analysis.go --test_database data_normalizer/testdata/s3stmtlog_1000

# example 2
xd -dk data_normalizer/testdata/s3stmtlog_100000.log.xz
sqlite3 s3stmtlog_100000 < data_normalizer/testdata/s3stmtlog_100000.sql
go run query_analysis.go --test_database data_normalizer/testdata/s3stmtlog_100000
```

## Sample result
The result from the sample test above is listed below. UNKNOWN represents the cases
we could not parse the SQL. Further work is needed to fix the logging code.
```
Query Type                         Frequency
INSERT_VALUES                      27294
SCAN                               12427
UNKNOWN                            3129
AGGREGATE_SCAN_NO_GROUPBY          1573
SYSTEM                             1253
AGGREGATE_SCAN_GROUPBY             673
AGGREGATE_JOIN_NO_GROUPBY          205
JOIN_NO_AGGREGATE                  178
AGGREGATE_JOIN_GROUPBY             126
UPDATE                             56
DELETE                             43
EXPLAIN                            1

Query Type                         Total Query Time in Seconds
INSERT_VALUES                      731208265312.00
AGGREGATE_SCAN_NO_GROUPBY          654343275078.00
UNKNOWN                            145661868391.00
SCAN                               86556227451.00
DELETE                             53140982570.00
AGGREGATE_SCAN_GROUPBY             37408239696.00
AGGREGATE_JOIN_GROUPBY             21250328873.00
AGGREGATE_JOIN_NO_GROUPBY          5766394256.00
JOIN_NO_AGGREGATE                  1503911970.00
SYSTEM                             504438257.00
UPDATE                             457730456.00
EXPLAIN                            3383641.00

Query Type                         Total MB Memory
INSERT_VALUES                      4220371801.00
DELETE                             963206076.00
SCAN                               350406301.00
UNKNOWN                            324431085.00
AGGREGATE_JOIN_GROUPBY             308020146.00
AGGREGATE_SCAN_GROUPBY             237963712.00
AGGREGATE_JOIN_NO_GROUPBY          232208935.00
JOIN_NO_AGGREGATE                  120207381.00
AGGREGATE_SCAN_NO_GROUPBY          69908415.00
SYSTEM                             37789696.00
UPDATE                             848408.00
EXPLAIN                            0.00

Query Type                         Frequency           Total Time          Total MB Memory     
Read                               15183               806831760965.00     1258.00             
Write                              31775               930973284986.00     5290.00             

Query Type                         Frequency           Total Time          Total MB Memory     
Insert Select                      0                   0.00                0.00                
Insert Values                      27294               731208265312.00     4025.00