package advisor

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pingcap/index_advisor/optimizer"
	"github.com/pingcap/index_advisor/utils"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
)

func isIndexableColumnType(tp *types.FieldType) bool {
	if tp == nil {
		return false
	}
	switch tp.Tp {
	case mysql.TypeTiny, mysql.TypeShort, mysql.TypeInt24, mysql.TypeLong, mysql.TypeLonglong, mysql.TypeYear,
		mysql.TypeFloat, mysql.TypeDouble, mysql.TypeNewDecimal,
		mysql.TypeDuration, mysql.TypeDate, mysql.TypeDatetime, mysql.TypeTimestamp:
		return true
	case mysql.TypeVarchar, mysql.TypeString, mysql.TypeVarString:
		return tp.Flen <= 512
	}
	return false
}

func evaluateIndexConfCostConcurrently(info utils.WorkloadInfo, optimizers []optimizer.WhatIfOptimizer,
	indexes []utils.Set[utils.Index]) (bestSet utils.Set[utils.Index], bestCost utils.IndexConfCost, err error) {
	bestSet = utils.NewSet[utils.Index]()
	errPointer := new(atomic.Pointer[error])
	costs := make([]utils.IndexConfCost, len(indexes))
	var wg sync.WaitGroup
	for id := 0; id < len(optimizers); id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := id; i < len(indexes); i += len(optimizers) {
				cost, err := evaluateIndexConfCost(info, optimizers[id], indexes[i])
				if err != nil {
					errPointer.CompareAndSwap(nil, &err)
					return
				}
				if errPointer.Load() != nil {
					return
				}
				costs[i] = cost
			}
		}(id)
	}
	wg.Wait()
	if errPointer.Load() != nil {
		return nil, bestCost, *errPointer.Load()
	}

	for i := 0; i < len(indexes); i++ {
		if costs[i].Less(bestCost) {
			bestSet = indexes[i]
			bestCost = costs[i]
		}
	}
	return bestSet, bestCost, nil
}

// evaluateIndexConfCost evaluates the workload cost under the given indexes.
func evaluateIndexConfCost(info utils.WorkloadInfo, optimizer optimizer.WhatIfOptimizer, indexes utils.Set[utils.Index]) (utils.IndexConfCost, error) {
	failedIndexes := utils.NewSet[utils.Index]()
	for _, index := range indexes.ToList() {
		if err := optimizer.CreateHypoIndex(index); err != nil {
			utils.Warningf("create %v failed: %v", index.Key(), err)
			failedIndexes.Add(index)
		}
	}
	var workloadCost float64
	for _, sql := range info.Queries.ToList() { // TODO: run them concurrently to save time
		p, err := optimizer.ExplainQ(sql)
		if err != nil {
			return utils.IndexConfCost{}, err
		}
		workloadCost += p.PlanCost() * float64(sql.Frequency)
	}
	for _, index := range indexes.ToList() {
		if failedIndexes.Contains(index) {
			continue
		}
		if err := optimizer.DropHypoIndex(index); err != nil {
			return utils.IndexConfCost{}, err
		}
	}
	var totCols int
	var keys []string
	for _, index := range indexes.ToList() {
		totCols += len(index.Columns)
		keys = append(keys, index.Key())
	}
	sort.Strings(keys)

	return utils.IndexConfCost{workloadCost, totCols, strings.Join(keys, ",")}, nil
}

var indexID atomic.Int64

// tempIndexName returns a temp index name for the given columns.
func tempIndexName(cols ...utils.Column) string {
	var names []string
	for _, col := range cols {
		names = append(names, col.ColumnName)
	}
	idxName := fmt.Sprintf("idx_%v", strings.Join(names, "_"))
	if len(idxName) <= 64 {
		return idxName
	}
	return fmt.Sprintf("idx_%v", indexID.Add(1))
}

func checkWorkloadInfo(w utils.WorkloadInfo) {
	for _, col := range w.IndexableColumns.ToList() {
		if col.SchemaName == "" || col.TableName == "" || col.ColumnName == "" {
			panic(fmt.Sprintf("invalid indexable column: %v", col))
		}
	}
	for _, sql := range w.Queries.ToList() {
		if sql.Text == "" {
			panic(fmt.Sprintf("invalid sql: %v", sql))
		}
		for _, col := range sql.IndexableColumns.ToList() {
			if col.SchemaName == "" || col.TableName == "" || col.ColumnName == "" {
				panic(fmt.Sprintf("invalid indexable column: %v", col))
			}
		}
	}
	for _, tbl := range w.TableSchemas.ToList() {
		if tbl.SchemaName == "" || tbl.TableName == "" {
			panic(fmt.Sprintf("invalid table schema: %v", tbl))
		}
		for _, col := range tbl.Columns {
			if col.SchemaName == "" || col.TableName == "" || col.ColumnName == "" || col.ColumnType == nil {
				panic(fmt.Sprintf("invalid indexable column: %v", col))
			}
		}
		for _, idx := range tbl.Indexes {
			if idx.SchemaName == "" || idx.TableName == "" || idx.IndexName == "" {
				panic(fmt.Sprintf("invalid index: %v", idx))
			}
			for _, col := range idx.Columns {
				if col.SchemaName == "" || col.TableName == "" || col.ColumnName == "" {
					panic(fmt.Sprintf("invalid indexable column: %v", col))
				}
			}
		}
	}
}
