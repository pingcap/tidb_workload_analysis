package advisor

import (
	"testing"

	"github.com/qw4990/index_advisor/utils"
)

func TestDigestCompression(t *testing.T) {
	s := utils.NewSet[utils.Query]()
	s.Add(utils.Query{Text: "select * from t1 where a = 1", Frequency: 1})
	s.Add(utils.Query{Text: "select * from t1 where a = 2", Frequency: 2})
	s.Add(utils.Query{Text: "select * from t1 where a = 3", Frequency: 3})
	cs := compressBySQLDigest(s)
	if cs.ToList()[0].Frequency != 1+2+3 {
		t.Errorf("expect 6, got %v", cs.ToList()[0].Frequency)
	}
}
