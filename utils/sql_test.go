package utils

import (
	"sort"
	"strings"
	"testing"
)

func TestNormalizeQueryWithDB(t *testing.T) {
	cases := []struct {
		q      string
		result string
	}{
		{`select * from t`, "select * from test.t"},
		{`select * from t1, t2 where t1.a<10`, "select * from test.t1, test.t2 where t1.a<10"},
		{`select * from t1, xxx.t2 where t1.a<10`, "select * from test.t1, xxx.t2 where t1.a<10"},
		{`select * from xxx.t1, t2 where t1.a<10`, "select * from xxx.t1, test.t2 where t1.a<10"},
	}

	for _, c := range cases {
		nq, err := NormalizeQueryWithDB(c.q, "test")
		if err != nil {
			t.Errorf("NormalizeQueryWithDB(%s) error: %v", c.q, err)
		}
		if nq != c.result {
			t.Errorf("NormalizeQueryWithDB(%s) = %s, expected %s", c.q, nq, c.result)
		}
	}
}

func TestParseOrderByColumnsFromQuery(t *testing.T) {
	cases := []struct {
		q string
		c []string
	}{
		{`select * from t where a = 1 or b = 2 or 3=c order by a, b, c`,
			[]string{"test.t.a", "test.t.b", "test.t.c"}},
		{`select * from t where a = 1 or b = 2 order by a, b`,
			[]string{"test.t.a", "test.t.b"}},
		{`select * from t where a = 1 order by a`,
			[]string{"test.t.a"}},
		{`select * from t where a = 1 and b =1 order by a, b`,
			[]string{"test.t.a", "test.t.b"}},
		{`select * from t where a = 1 and (b =1 or c=1) order by a, b, c`,
			[]string{"test.t.a", "test.t.b", "test.t.c"}},
		// unsupported
		{`select * from t1, t2 where b =1 or c=1 order by a, b, c`,
			[]string{}},
		{`select * from t where a = 1 and b =1 order by a+1, b`,
			[]string{}},
	}

	for _, c := range cases {
		result, err := ParseOrderByColumnsFromQuery(Query{
			SchemaName: "test",
			Text:       c.q,
		})
		must(err)

		var getColStrs []string
		for _, col := range result {
			getColStrs = append(getColStrs, col.Key())
		}
		get := strings.Join(getColStrs, ",")
		expected := strings.Join(c.c, ",")
		if get != expected {
			t.Errorf("ParseOrderByColumnsFromQuery(%s) = %s, expected %s", c.q, get, expected)
		}
	}
}

func TestParseSelectColumnsFromQuery(t *testing.T) {
	cases := []struct {
		q string
		c []string
	}{
		{`select a, b, c from t where a = 1 or b = 2 or 3=c`,
			[]string{"test.t.a", "test.t.b", "test.t.c"}},
		{`select a, b from t where a = 1 or b = 2`,
			[]string{"test.t.a", "test.t.b"}},
		{`select a from t where b = 1`,
			[]string{"test.t.a"}},
		{`select * from t where a = 1`,
			[]string{}},
		{`select c from t where a = 1 and b =1`,
			[]string{"test.t.c"}},
		{`select b, c from t where a = 1 and (b =1 or c=1)`,
			[]string{"test.t.b", "test.t.c"}},
		{`select * from t1, t2 where b =1 or c=1`,
			[]string{}}, // unsupported
	}

	for _, c := range cases {
		result, err := ParseSelectColumnsFromQuery(Query{
			SchemaName: "test",
			Text:       c.q,
		})
		must(err)
		checkDNFColResult(result, c.c)
	}
}

func TestParseDNFColumnsFromQuery(t *testing.T) {
	cases := []struct {
		q string
		c []string
	}{
		{`select * from t where a = 1 or b = 2 or 3=c`,
			[]string{"test.t.a", "test.t.b", "test.t.c"}},
		{`select * from t where a = 1 or b = 2`,
			[]string{"test.t.a", "test.t.b"}},
		{`select * from t where a = 1`,
			[]string{}},
		{`select * from t where a = 1 and b =1`,
			[]string{}},
		{`select * from t where a = 1 and (b =1 or c=1)`,
			[]string{"test.t.b", "test.t.c"}},
		{`select * from t1, t2 where b =1 or c=1`,
			[]string{}}, // unsupported
		{
			`SELECT * FROM t WHERE
  				timestamp >= 1647469098 AND timestamp <= 1679005097
  				AND ( from_address = "eth:aaa" OR to_address = "eth:bbb" )
  				AND from_address <> to_address
				ORDER BY timestamp DESC LIMIT 200`,
			[]string{"test.t.from_address", "test.t.to_address"},
		},
		{
			`SELECT * FROM t WHERE
				  1 = 1
				  AND ( from_address = "aaa" OR to_address = "bbb" )
				  AND timestamp <= 1676423335681
				LIMIT 9007199254740991`,
			[]string{"test.t.from_address", "test.t.to_address"},
		},
	}

	for _, c := range cases {
		result, err := ParseDNFColumnsFromQuery(Query{
			SchemaName: "test",
			Text:       c.q,
		})
		must(err)
		checkDNFColResult(result, c.c)
	}
}

func checkDNFColResult(got Set[Column], want []string) {
	var gotStr []string
	if got != nil {
		for _, c := range got.ToList() {
			gotStr = append(gotStr, c.Key())
		}
	}
	sort.Strings(gotStr)
	sort.Strings(want)

	kGot := strings.Join(gotStr, ",")
	kWant := strings.Join(want, ",")
	if kGot != kWant {
		panic("got: " + kGot + ", want: " + kWant)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
