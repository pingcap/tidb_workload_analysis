# Query Analysis (Beta)

This tool analyzes the structure of workload queries stored in a SQLite database.

## Prerequisites

- Go and a C compiler (the SQLite driver uses cgo).
- A SQLite database containing a `statements_summary` table with the columns
  used by the query in `query_analysis.go`. Provide your own data; the repository
  no longer includes workload samples.

## Run

From this directory, run:

```sh
go run query_analysis.go --test_database /path/to/querylog_database
```

The tool drops and recreates the `unique_queries` table in the input database.
Use a copy if the existing table must be preserved. The table contains sample
SQL, so protect the database and any reports according to the sensitivity of
the workload data.
