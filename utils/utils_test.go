package utils

import (
	"fmt"
	"testing"
)

func TestCollectTableNames(t *testing.T) {
	sql := `
SELECT MIN(mc.note) AS production_note, MIN(t.title) AS movie_title
	, MIN(t.production_year) AS movie_year
FROM company_type ct, info_type it, movie_companies mc, movie_info_idx mi_idx, title t
WHERE ct.kind = 'production companies'
	AND it.info = 'top 250 rank'
	AND mc.note NOT LIKE '%(as Metro-Goldwyn-Mayer Pictures)%'
	AND (mc.note LIKE '%(co-production)%'
		OR mc.note LIKE '%(presents)%')
	AND ct.id = mc.company_type_id
	AND t.id = mc.movie_id
	AND t.id = mi_idx.movie_id
	AND mc.movie_id = mi_idx.movie_id
	AND it.id = mi_idx.info_type_id;
`
	tables, _ := CollectTableNamesFromSQL("test", sql)
	fmt.Println(tables.ToList())
}

func TestCombSet(t *testing.T) {
	s := NewSet[Column]()
	for i := 0; i < 6; i++ {
		s.Add(NewColumn("test", "test", fmt.Sprintf("col%d", i)))
	}

	for i := 1; i < 6; i++ {
		fmt.Println("======================== ", i, " ========================")
		result := CombSet(s, i)
		fmt.Println("--> ", len(result))
		for _, ss := range result {
			fmt.Println(ss.ToList())
		}
	}
}

func TestPlanCost(t *testing.T) {
	plan := [][]string{
		{"HashJoin_37", "100", "8225.40"},
		{"├─IndexHashJoin_45(Build)", "1.000", "6096.63"},
		{"└─CTEFullScan_39(Probe)", "10.00", "14.97"},
		{"CTE_0", "10.00", "14.97"},
		{"└─IndexLookUp_31(Seed Part)", "10.00", "19530.45"},
	}
	p := Plan(plan)
	if p.PlanCost() != 8225.40+19530.45 {
		t.Error("plan cost error")
	}
}
