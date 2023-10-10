package advisor

import (
	"github.com/qw4990/index_advisor/optimizer"
	"github.com/qw4990/index_advisor/utils"
)

// IndexSelectionAlgo is the interface for index selection algorithms.
type IndexSelectionAlgo func(
	workloadInfo utils.WorkloadInfo, // the target workload
	parameter Parameter, // the input parameters
	optimizer optimizer.WhatIfOptimizer, // the what-if optimizer
) (utils.Set[utils.Index], error)

// IndexableColumnsSelectionAlgo is the interface for indexable columns selection algorithms.
type IndexableColumnsSelectionAlgo func(workloadInfo *utils.WorkloadInfo) error

// WorkloadInfoCompressionAlgo is the interface for workload info compression algorithms.
type WorkloadInfoCompressionAlgo func(workloadInfo utils.WorkloadInfo) utils.WorkloadInfo

var (
	compressAlgorithms = map[string]WorkloadInfoCompressionAlgo{
		"none":   NoneWorkloadInfoCompress,
		"digest": DigestWorkloadInfoCompress,
	}

	findIndexableColsAlgorithms = map[string]IndexableColumnsSelectionAlgo{
		"simple": IndexableColumnsSelectionSimple,
	}

	selectIndexAlgorithms = map[string]IndexSelectionAlgo{
		"auto_admin": SelectIndexAAAlgo,
	}
)

// Parameter is the input parameters of index advisor.
type Parameter struct {
	MaxNumberIndexes int // the max number of indexes to recommend
	MaxIndexWidth    int // the max number of columns in recommended indexes
}

func validateParameter(p Parameter) Parameter {
	if p.MaxNumberIndexes < 1 {
		utils.Warningf("max number of indexes should be at least 1, set from %v to 1", p.MaxNumberIndexes)
		p.MaxNumberIndexes = 1
	}
	if p.MaxNumberIndexes > 20 {
		utils.Warningf("max number of indexes should be at most 20, set from %v to 20", p.MaxNumberIndexes)
		p.MaxNumberIndexes = 20
	}
	if p.MaxIndexWidth < 1 {
		utils.Warningf("max index width should be at least 1, set from %v to 1", p.MaxIndexWidth)
		p.MaxIndexWidth = 1
	}
	if p.MaxIndexWidth > 5 {
		utils.Warningf("max index width should be at most 5, set from %v to 5", p.MaxIndexWidth)
		p.MaxIndexWidth = 5
	}
	return p
}

// IndexAdvise is the entry point of index advisor.
func IndexAdvise(db optimizer.WhatIfOptimizer, workload utils.WorkloadInfo, param Parameter) (utils.Set[utils.Index], error) {
	utils.Infof("start index advise for %v queries, %v tables", workload.Queries.Size(), workload.TableSchemas.Size())
	param = validateParameter(param)

	compress := compressAlgorithms["digest"]
	indexable := findIndexableColsAlgorithms["simple"]
	selection := selectIndexAlgorithms["auto_admin"]

	compressedWorkloadInfo := compress(workload)
	utils.Infof("compress %v queries to %v queries", workload.Queries.Size(), compressedWorkloadInfo.Queries.Size())

	if err := indexable(&compressedWorkloadInfo); err != nil {
		return nil, err
	}
	utils.Infof("find %v indexable columns", compressedWorkloadInfo.IndexableColumns.Size())

	checkWorkloadInfo(compressedWorkloadInfo)
	recommendedIndexes, err := selection(compressedWorkloadInfo, param, db)
	if err != nil {
		return nil, err
	}
	utils.Infof("finish index advise with %v recommended indexes", recommendedIndexes.Size())
	return recommendedIndexes, err
}
