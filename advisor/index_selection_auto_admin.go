package advisor

import (
	"github.com/pingcap/index_advisor/optimizer"
	"github.com/pingcap/index_advisor/utils"
)

/*
	This algorithm resembles the index selection algorithm published in 1997 by Chaudhuri
	and Narasayya. Details can be found in the original paper:
	Surajit Chaudhuri, Vivek R. Narasayya: An Efficient Cost-Driven Index Selection
	Tool for Microsoft Query Server. VLDB 1997: 146-155
	This implementation is the Golang version of github.com/hyrise/index_selection_evaluation/blob/refactoring/selection/algorithms/auto_admin_algorithm.py.
*/

// SelectIndexAAAlgo implements the auto-admin algorithm.
func SelectIndexAAAlgo(workload utils.WorkloadInfo, parameter Parameter, op optimizer.WhatIfOptimizer) (utils.Set[utils.Index], error) {
	concurrency := 8 // TODO: make it configurable
	tmpOptimizers := make([]optimizer.WhatIfOptimizer, concurrency)
	for i := 0; i < concurrency; i++ {
		var err error
		tmpOptimizers[i], err = op.Clone()
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		for _, tOpt := range tmpOptimizers {
			if tOpt != nil {
				tOpt.Close()
			}
		}
	}()

	aa := &autoAdmin{
		optimizer:     op,
		tmpOptimizers: tmpOptimizers,
		maxIndexes:    parameter.MaxNumberIndexes,
		maxIndexWidth: parameter.MaxIndexWidth,
	}
	utils.Infof("starting auto-admin algorithm with max-indexes %d, max index-width %d", aa.maxIndexes, aa.maxIndexWidth)

	op.ResetStats()
	bestIndexes, err := aa.calculateBestIndexes(workload)
	if err != nil {
		return nil, err
	}
	utils.Infof("what-if optimizer stats: %v", op.Stats().Format())
	return bestIndexes, nil
}

type autoAdmin struct {
	optimizer     optimizer.WhatIfOptimizer
	tmpOptimizers []optimizer.WhatIfOptimizer // used to run SQLs concurrently

	maxIndexes    int // The algorithm stops as soon as it has selected #max_indexes indexes
	maxIndexWidth int // The number of columns an index can contain at maximum.
}

func (aa *autoAdmin) calculateBestIndexes(workload utils.WorkloadInfo) (utils.Set[utils.Index], error) {
	if aa.maxIndexes == 0 {
		return nil, nil
	}

	potentialIndexes := utils.NewSet[utils.Index]() // each indexable column as a single-column index
	for _, col := range workload.IndexableColumns.ToList() {
		potentialIndexes.Add(utils.NewIndex(col.SchemaName, col.TableName, tempIndexName(col), col.ColumnName))
	}

	currentBestIndexes := utils.NewSet[utils.Index]()
	for currentMaxIndexWidth := 1; currentMaxIndexWidth <= aa.maxIndexWidth; currentMaxIndexWidth++ {
		utils.Infof("auto-admin algorithm: current index width is %d", currentMaxIndexWidth)
		candidates, err := aa.selectIndexCandidates(workload, potentialIndexes)
		if err != nil {
			return nil, err
		}

		//maxIndexes := aa.maxIndexes * (aa.maxIndexWidth - currentMaxIndexWidth + 1)
		maxIndexes := aa.maxIndexes
		utils.Infof("auto-admin algorithm: select best %v candidate indexes from %v candidates", maxIndexes, candidates.Size())
		currentBestIndexes, err = aa.enumerateCombinations(workload, candidates, maxIndexes)
		if err != nil {
			return nil, err
		}
		utils.Infof("auto-admin algorithm: select %v best candidate indexes", currentBestIndexes.Size())

		if currentMaxIndexWidth < aa.maxIndexWidth {
			// Update potential indexes for the next iteration
			potentialIndexes = currentBestIndexes
			potentialIndexes.AddSet(aa.createMultiColumnIndexes(workload, currentBestIndexes))
		}
	}

	currentBestIndexes, err := aa.heuristicMergeIndexes(currentBestIndexes, workload)
	if err != nil {
		return nil, err
	}
	currentBestIndexes, err = aa.heuristicCoveredIndexes(currentBestIndexes, workload)
	if err != nil {
		return nil, err
	}

	utils.Infof("auto-admin algorithm: the number of candidate indexes before filter is %v", currentBestIndexes.Size())
	currentBestIndexes, err = aa.filterIndexes(workload, currentBestIndexes)
	if err != nil {
		return nil, err
	}

	currentBestIndexes, err = aa.cutDown(currentBestIndexes, workload, aa.optimizer, aa.maxIndexes)
	if err != nil {
		return nil, err
	}

	// try to add more indexes if the number of indexes is less than maxIndexes
	for limit := 0; limit < 3 && currentBestIndexes.Size() < aa.maxIndexes; limit++ {
		potentialIndexes = utils.DiffSet(potentialIndexes, currentBestIndexes)
		currentCost, err := evaluateIndexConfCost(workload, aa.optimizer, currentBestIndexes)
		if err != nil {
			return nil, err
		}
		currentBestIndexes, _, err = aa.enumerateGreedy(workload, currentBestIndexes, currentCost, potentialIndexes, aa.maxIndexes)
		if err != nil {
			return nil, err
		}

		currentBestIndexes, err = aa.filterIndexes(workload, currentBestIndexes)
		if err != nil {
			return nil, err
		}
	}

	return currentBestIndexes, nil
}

// cutDown removes indexes from candidateIndexes until the number of indexes is less than or equal to maxIndexes.
func (aa *autoAdmin) cutDown(candidateIndexes utils.Set[utils.Index],
	w utils.WorkloadInfo, op optimizer.WhatIfOptimizer, maxIndexes int) (utils.Set[utils.Index], error) {
	if candidateIndexes.Size() <= maxIndexes {
		return candidateIndexes, nil
	}

	// find the target index to remove, which is the one that has the least impact on the cost.
	var bestCost utils.IndexConfCost
	var targetIndex utils.Index
	for i, idx := range candidateIndexes.ToList() {
		candidateIndexes.Remove(idx)
		cost, err := evaluateIndexConfCost(w, op, candidateIndexes)
		if err != nil {
			return nil, err
		}
		candidateIndexes.Add(idx)

		if i == 0 || cost.Less(bestCost) {
			bestCost = cost
			targetIndex = idx
		}
	}

	candidateIndexes.Remove(targetIndex)
	return aa.cutDown(candidateIndexes, w, op, maxIndexes)
}

func (aa *autoAdmin) heuristicCoveredIndexes(candidateIndexes utils.Set[utils.Index], w utils.WorkloadInfo) (utils.Set[utils.Index], error) {
	// build an index (b, a) for `select a from t where b=1` to convert IndexLookup to IndexScan
	currentCost, err := evaluateIndexConfCost(w, aa.optimizer, candidateIndexes)
	if err != nil {
		return nil, err
	}

	for _, q := range w.Queries.ToList() {
		// parse select columns
		selectCols, err := utils.ParseSelectColumnsFromQuery(q)
		if err != nil {
			return nil, err
		}
		if selectCols == nil || selectCols.Size() == 0 || selectCols.Size() > aa.maxIndexWidth {
			continue
		}
		schemaName, tableName := selectCols.ToList()[0].SchemaName, selectCols.ToList()[0].TableName

		// generate cover-index candidates
		coverIndexSet := utils.NewSet[utils.Index]()
		coverIndexSet.Add(utils.Index{
			SchemaName: schemaName,
			TableName:  tableName,
			IndexName:  tempIndexName(selectCols.ToList()...),
			Columns:    selectCols.ToList(),
		})
		for _, idx := range candidateIndexes.ToList() {
			if idx.SchemaName != schemaName || idx.TableName != tableName {
				continue // not for the same table
			}
			if len(idx.Columns)+selectCols.Size() > aa.maxIndexWidth {
				continue // exceed the max-index-width limitation
			}
			// try this cover-index: idx-cols + select-cols
			var newCols []utils.Column
			for _, col := range selectCols.ToList() {
				duplicated := false
				for _, idxCol := range idx.Columns {
					if col.Key() == idxCol.Key() {
						duplicated = true
						break
					}
				}
				if !duplicated {
					newCols = append(newCols, col)
				}
			}
			var cols []utils.Column
			cols = append(cols, idx.Columns...)
			cols = append(cols, newCols...)
			coverIndexSet.Add(utils.Index{
				SchemaName: schemaName,
				TableName:  tableName,
				IndexName:  tempIndexName(cols...),
				Columns:    cols,
			})
		}

		// select the best cover-index
		var bestCoverIndex utils.Index
		var bestCoverIndexCost utils.IndexConfCost
		for i, coverIndex := range coverIndexSet.ToList() {
			candidateIndexes.Add(coverIndex)
			cost, err := evaluateIndexConfCost(w, aa.optimizer, candidateIndexes)
			if err != nil {
				return nil, err
			}
			candidateIndexes.Remove(coverIndex)

			if i == 0 || cost.Less(bestCoverIndexCost) {
				bestCoverIndexCost = cost
				bestCoverIndex = coverIndex
			}
		}

		// check whether this cover-index can bring any benefits
		if bestCoverIndexCost.Less(currentCost) {
			candidateIndexes.Add(bestCoverIndex)
			currentCost = bestCoverIndexCost
		}
	}

	return candidateIndexes, nil
}

func (aa *autoAdmin) heuristicMergeIndexes(candidateIndexes utils.Set[utils.Index], w utils.WorkloadInfo) (utils.Set[utils.Index], error) {
	// try to build index set {(c1), (c2)} for predicate like `where c1=1 or c2=2` so that index-merge can be applied.
	currentCost, err := evaluateIndexConfCost(w, aa.optimizer, candidateIndexes)
	if err != nil {
		return nil, err
	}

	for _, q := range w.Queries.ToList() {
		// get all DNF columns from the query
		dnfCols, err := utils.ParseDNFColumnsFromQuery(q)
		if err != nil {
			return nil, err
		}
		if dnfCols == nil || dnfCols.Size() == 0 {
			continue
		}
		orderByCols, err := utils.ParseOrderByColumnsFromQuery(q)
		if err != nil {
			return nil, err
		}

		// create indexes for these DNF columns
		newIndexes := utils.NewSet[utils.Index]()
		for _, col := range dnfCols.ToList() {
			idx := utils.NewIndex(col.SchemaName, col.TableName, tempIndexName(col), col.ColumnName)
			contained := false
			for _, existingIndex := range candidateIndexes.ToList() {
				if existingIndex.PrefixContain(idx) {
					contained = true
					continue
				}
			}
			if !contained {
				newIndexes.Add(idx)
			}

			// index with DNF column + order-by column
			if len(orderByCols) == 0 {
				continue
			}
			cols := []utils.Column{col}
			cols = append(cols, orderByCols...)
			if len(cols) > aa.maxIndexWidth {
				cols = cols[:aa.maxIndexWidth]
			}
			idx = utils.NewIndexWithColumns(tempIndexName(cols...), cols...)
			contained = false
			for _, existingIndex := range candidateIndexes.ToList() {
				if existingIndex.PrefixContain(idx) {
					contained = true
					continue
				}
			}
			if !contained {
				newIndexes.Add(idx)
			}
		}
		if newIndexes.Size() == 0 {
			continue
		}

		// check whether these new indexes for IndexMerge can bring some benefits.
		newCandidateIndexes := utils.UnionSet(candidateIndexes, newIndexes)
		newCost, err := evaluateIndexConfCost(w, aa.optimizer, newCandidateIndexes)
		if err != nil {
			return nil, err
		}
		if newCost.Less(currentCost) {
			currentCost = newCost
			candidateIndexes, err = aa.filterIndexes(w, newCandidateIndexes)
			if err != nil {
				return nil, err
			}
		}
	}

	return candidateIndexes, nil
}

func (aa *autoAdmin) createMultiColumnIndexes(workload utils.WorkloadInfo, indexes utils.Set[utils.Index]) utils.Set[utils.Index] {
	multiColumnCandidates := utils.NewSet[utils.Index]()
	for _, index := range indexes.ToList() {
		table, found := workload.TableSchemas.Find(utils.TableSchema{SchemaName: index.SchemaName, TableName: index.TableName})
		if !found {
			continue
		}
		tableColsSet := utils.ListToSet[utils.Column](table.Columns...)
		indexableColsSet := workload.IndexableColumns
		indexColsSet := utils.ListToSet[utils.Column](index.Columns...)
		for _, column := range utils.DiffSet(utils.AndSet(tableColsSet, indexableColsSet), indexColsSet).ToList() {
			cols := append([]utils.Column{}, index.Columns...)
			cols = append(cols, column)
			multiColumnCandidates.Add(utils.Index{
				SchemaName: index.SchemaName,
				TableName:  index.TableName,
				IndexName:  tempIndexName(cols...),
				Columns:    cols,
			})
		}
	}
	return multiColumnCandidates
}

// filterIndexes filters some obviously unreasonable indexes.
// Rule 1: if index X is a prefix of index Y, then remove X.
// Rule 2: if index X has no any benefit, then remove X.
// Rule 3: if candidate index X is a prefix of some existing index in the workload, then remove X.
// Rule 4(TBD): remove unnecessary suffix columns, e.g. X(a, b, c) to X(a, b) if no query can gain benefit from the suffix column c.
func (aa *autoAdmin) filterIndexes(workload utils.WorkloadInfo, indexes utils.Set[utils.Index]) (utils.Set[utils.Index], error) {
	indexList := indexes.ToList()
	filteredIndexes := utils.NewSet[utils.Index]()
	originalCost, err := evaluateIndexConfCost(workload, aa.optimizer, indexes)
	if err != nil {
		return nil, err
	}
	for i, x := range indexList {
		filtered := false
		// rule 1
		for j, y := range indexList {
			if i == j {
				continue
			}
			if y.PrefixContain(x) {
				filtered = true
				continue
			}
		}
		if filtered {
			continue
		}

		// rule 2
		indexes.Remove(x)
		newCost, err := evaluateIndexConfCost(workload, aa.optimizer, indexes)
		if err != nil {
			return nil, err
		}
		indexes.Add(x)
		if !originalCost.Less(newCost) {
			continue
		}

		// rule 3
		table, ok := workload.TableSchemas.Find(utils.TableSchema{SchemaName: x.SchemaName, TableName: x.TableName})
		if ok {
			prefixContain := false
			for _, existingIndex := range table.Indexes {
				if existingIndex.PrefixContain(x) {
					prefixContain = true
				}
			}
			if prefixContain {
				continue
			}
		}

		filteredIndexes.Add(x)
	}
	return filteredIndexes, nil
}

// selectIndexCandidates selects the best indexes for each single-query.
func (aa *autoAdmin) selectIndexCandidates(workload utils.WorkloadInfo, potentialIndexes utils.Set[utils.Index]) (utils.Set[utils.Index], error) {
	utils.Debugf("auto-admin algorithm: select best index from %v candidates for each single query(total %v)", potentialIndexes.Size(), workload.Queries.Size())
	candidates := utils.NewSet[utils.Index]()
	for _, query := range workload.Queries.ToList() {
		queryWorkload := utils.WorkloadInfo{ // each query as a workload
			Queries:      utils.ListToSet(query),
			TableSchemas: workload.TableSchemas,
			TableStats:   workload.TableStats,
		}
		indexes := aa.potentialIndexesForQuery(query, potentialIndexes)

		bestPerQuery := 3 // keep 3 best indexes for each single-query
		bestQueryIndexes := utils.NewSet[utils.Index]()
		for i := 0; i < bestPerQuery; i++ {
			best, err := aa.enumerateCombinations(queryWorkload, indexes, 1)
			if err != nil {
				return nil, err
			}
			if best.Size() == 0 {
				break
			}
			bestQueryIndexes.AddSet(best)
			for _, index := range best.ToList() {
				indexes.Remove(index)
			}
			if bestQueryIndexes.Size() > bestPerQuery {
				break
			}
		}

		utils.Debugf("auto-admin algorithm: current best index for %s is %s", query.Alias, bestQueryIndexes.ToKeyList())
		candidates.AddSet(bestQueryIndexes)
	}
	return candidates, nil
}

// potentialIndexesForQuery returns best recommended indexes of this workload from these candidates.
func (aa *autoAdmin) enumerateCombinations(workload utils.WorkloadInfo,
	candidateIndexes utils.Set[utils.Index],
	maxNumberIndexes int) (utils.Set[utils.Index], error) {
	maxIndexesNative := 2
	if candidateIndexes.Size() > 50 {
		maxIndexesNative = 1
	}
	numberIndexesNaive := utils.Min(maxIndexesNative, candidateIndexes.Size(), maxNumberIndexes)
	currentIndexes, cost, err := aa.enumerateNaive(workload, candidateIndexes, numberIndexesNaive)
	if err != nil {
		return nil, err
	}

	numberIndexes := utils.Min(maxNumberIndexes, candidateIndexes.Size())
	indexes, cost, err := aa.enumerateGreedy(workload, currentIndexes, cost, candidateIndexes, numberIndexes)
	return indexes, err
}

// enumerateGreedy finds the best combination of indexes with a greedy algorithm.
func (aa *autoAdmin) enumerateGreedy(workload utils.WorkloadInfo, currentIndexes utils.Set[utils.Index],
	currentCost utils.IndexConfCost, candidateIndexes utils.Set[utils.Index], numberIndexes int) (utils.Set[utils.Index], utils.IndexConfCost, error) {
	if currentIndexes.Size() >= numberIndexes || candidateIndexes.Size() == 0 {
		return currentIndexes, currentCost, nil
	}

	// iterate all unused indexes and add one into the current set
	indexCombinations := make([]utils.Set[utils.Index], 0, 128)
	for _, index := range candidateIndexes.ToList() {
		newCombination := utils.UnionSet(currentIndexes, utils.ListToSet(index))
		if newCombination.Size() != currentIndexes.Size()+1 {
			continue // duplicated index
		}
		indexCombinations = append(indexCombinations, newCombination)
	}
	if len(indexCombinations) == 0 {
		return currentIndexes, currentCost, nil
	}

	// find the best set
	bestSet, bestCost, err := evaluateIndexConfCostConcurrently(workload, aa.tmpOptimizers, indexCombinations)
	if err != nil {
		return nil, bestCost, err
	}
	if bestSet.Size() == 0 {
		return currentIndexes, currentCost, nil
	}
	bestNewIndex := utils.DiffSet(bestSet, currentIndexes).ToList()[0]
	if bestCost.Less(currentCost) {
		currentIndexes.Add(bestNewIndex)
		candidateIndexes.Remove(bestNewIndex)
		currentCost = bestCost
		return aa.enumerateGreedy(workload, currentIndexes, currentCost, candidateIndexes, numberIndexes)
	}

	return currentIndexes, currentCost, nil
}

// enumerateNaive enumerates all possible combinations of indexes with at most numberIndexesNaive indexes and returns the best one.
func (aa *autoAdmin) enumerateNaive(workload utils.WorkloadInfo, candidateIndexes utils.Set[utils.Index], numberIndexesNaive int) (utils.Set[utils.Index], utils.IndexConfCost, error) {
	// get all index combinations
	indexCombinations := make([]utils.Set[utils.Index], 0, 128)
	for numberOfIndexes := 1; numberOfIndexes <= numberIndexesNaive; numberOfIndexes++ {
		indexCombinations = append(indexCombinations, utils.CombSet(candidateIndexes, numberOfIndexes)...)
	}
	if len(indexCombinations) > 32 {
		utils.Infof("auto-admin algorithm: find %v index combinations", len(indexCombinations))
	}

	lowestCostIndexes, lowestCost, err := evaluateIndexConfCostConcurrently(workload, aa.tmpOptimizers, indexCombinations)
	if err != nil {
		return nil, lowestCost, err
	}
	if len(indexCombinations) > 32 {
		utils.Infof("auto-admin algorithm: find the best combination from %v combinations", len(indexCombinations))
	}

	return lowestCostIndexes, lowestCost, nil
}

func (aa *autoAdmin) potentialIndexesForQuery(query utils.Query, potentialIndexes utils.Set[utils.Index]) utils.Set[utils.Index] {
	indexes := utils.NewSet[utils.Index]()
	for _, index := range potentialIndexes.ToList() {
		// The leading index column must be referenced by the query.
		if query.IndexableColumns.Contains(index.Columns[0]) {
			indexes.Add(index)
		}
	}
	return indexes
}

//// mergeCandidates merges some index candidates.
//// Rule 1: if candidate index X has no benefit, then remove X.
//// Rule 2: if candidate index X is a prefix of some existing index in the workload, then remove X.
//// Rule 3(TBD): if candidate index X is a prefix of another candidate Y and Y's workload cost is less than X's, then remove X.
//func (aa *autoAdmin) mergeCandidates(workload utils.WorkloadInfo, candidates utils.Set[utils.Index]) (utils.Set[utils.Index], error) {
//	mergedCandidates := utils.NewSet[utils.Index]()
//	candidatesList := candidates.ToList()
//	var candidateCosts []utils.IndexConfCost
//	for _, c := range candidatesList {
//		cost, err := evaluateIndexConfCost(workload, aa.optimizer, utils.ListToSet(c))
//		if err != nil {
//			return nil, err
//		}
//		candidateCosts = append(candidateCosts, cost)
//	}
//	originalCost, err := evaluateIndexConfCost(workload, aa.optimizer, utils.NewSet[utils.Index]())
//	if err != nil {
//		return nil, err
//	}
//	for i, x := range candidatesList {
//		// rule 1
//		if originalCost.Less(candidateCosts[i]) {
//			continue
//		}
//
//		// rule 2
//		table, ok := workload.TableSchemas.Find(utils.TableSchema{SchemaName: x.SchemaName, TableName: x.TableName})
//		if !ok {
//			panic("table not found")
//		}
//		for _, existingIndex := range table.Indexes {
//			if existingIndex.PrefixContain(x) {
//				continue
//			}
//		}
//
//		//// rule 3
//		//hitRule3 := false
//		//for j, y := range candidatesList {
//		//	if i == j {
//		//		continue
//		//	}
//		//	// X is a prefix of Y and Y's cost is less than X's
//		//	if y.PrefixContain(x) && candidateCosts[j].Less(candidateCosts[i]) {
//		//		hitRule3 = true
//		//		break
//		//	}
//		//}
//		//if hitRule3 {
//		//	continue
//		//}
//		mergedCandidates.Add(x)
//	}
//	return mergedCandidates, nil
//}
