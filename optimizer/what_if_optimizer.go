package optimizer

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/qw4990/index_advisor/utils"
)

// WhatIfOptimizerStats records the statistics of a what-if optimizer.
type WhatIfOptimizerStats struct {
	ExecuteCount             int           // number of executed Query statements
	ExecuteTime              time.Duration // total execution time
	CreateOrDropHypoIdxCount int           // number of executed CreateHypoIndex/DropHypoIndex
	CreateOrDropHypoIdxTime  time.Duration // total execution time of CreateHypoIndex/DropHypoIndex
	GetCostCount             int           // number of executed GetCost
	GetCostTime              time.Duration // total execution time of GetCost
}

// Format formats the statistics.
func (s WhatIfOptimizerStats) Format() string {
	return fmt.Sprintf(`Execute(count/time): (%v/%v), CreateOrDropHypoIndex: (%v/%v), GetCost: (%v/%v)`,
		s.ExecuteCount, s.ExecuteTime, s.CreateOrDropHypoIdxCount, s.CreateOrDropHypoIdxTime, s.GetCostCount, s.GetCostTime)
}

// WhatIfOptimizer is the interface of a what-if optimizer.
type WhatIfOptimizer interface {
	Query(sql string) (*sql.Rows, error) // execute the specified Query statement and return the result
	Execute(sql string) error            // execute the specified Query statement

	Close() error                    // release the underlying database connection
	Clone() (WhatIfOptimizer, error) // clone this optimizer

	CreateHypoIndex(index utils.Index) error // create a hypothetical index
	DropHypoIndex(index utils.Index) error   // drop a hypothetical index

	ExplainQ(q utils.Query) (plan utils.Plan, err error)      // return the execution plan of the specified query
	Explain(query string) (plan utils.Plan, err error)        // return the execution plan of the specified query
	ExplainAnalyze(query string) (plan utils.Plan, err error) // return the execution plan of the specified query with analyze

	ResetStats()                 // reset the statistics
	Stats() WhatIfOptimizerStats // return the statistics

	SetDebug(flag bool) // print each query if set to true
}
