package utils

import (
	"github.com/pingcap/parser/format"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/opcode"
	_ "github.com/pingcap/tidb/types/parser_driver"
	driver "github.com/pingcap/tidb/types/parser_driver"
)

type StmtType int

const (
	StmtCreateDB StmtType = iota
	StmtUseDB
	StmtCreateTable
	StmtCreateIndex
	StmtSelect
	StmtUnknown
)

// GetStmtType returns the type of the given statement.
func GetStmtType(stmt string) StmtType {
	containAll := func(s string, substrs ...string) bool {
		s = strings.ToLower(s)
		for _, substr := range substrs {
			if !strings.Contains(s, substr) {
				return false
			}
		}
		return true
	}

	if containAll(stmt, "create", "database") {
		return StmtCreateDB
	} else if containAll(stmt, "create", "table") {
		return StmtCreateTable
	} else if containAll(stmt, "create", "index") {
		return StmtCreateIndex
	} else if containAll(stmt, "select", "from") {
		return StmtSelect
	} else if containAll(stmt, "use") {
		return StmtUseDB
	}
	return StmtUnknown
}

// GetDBNameFromUseDBStmt returns the database name of the given `USE` statement.
func GetDBNameFromUseDBStmt(stmt string) string {
	db := strings.Split(stmt, " ")[1]
	db = strings.Trim(db, "` '\"")
	return db
}

// GetDBNameFromCreateDBStmt returns the database name of the given `CREATE DATABASE` statement.
func GetDBNameFromCreateDBStmt(stmt string) string {
	tmp := strings.Split(stmt, " ")
	db := tmp[len(tmp)-1]
	db = strings.Trim(db, "` '\"")
	return db
}

// ParseOneSQL parses the given Query text and returns the AST.
func ParseOneSQL(sqlText string) (ast.StmtNode, error) {
	p := parser.New()
	return p.ParseOneStmt(sqlText, "", "")
}

// NormalizeDigest normalizes the given Query text and returns the normalized Query text and its digest.
func NormalizeDigest(sqlText string) (string, string) {
	return parser.NormalizeDigest(sqlText)
}

// NormalizeQueryWithDB `select * from t` --> `select * from db.t`
func NormalizeQueryWithDB(queryText, db string) (normalizedText string, err error) {
	stmt, err := ParseOneSQL(queryText)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	ctx := format.NewRestoreCtx(format.RestoreStringSingleQuotes|format.RestoreKeyWordLowercase|format.RestoreSpacesAroundBinaryOperation|format.RestoreStringWithoutCharset, &sb)
	ctx.DefaultDB = db
	if err := stmt.Restore(ctx); err != nil {
		return "", err
	}
	return sb.String(), nil
}

type tableNameCollector struct {
	defaultSchemaName string
	tableNames        Set[TableName]
	cteNames          Set[TableName]
}

func (c *tableNameCollector) Enter(n ast.Node) (out ast.Node, skipChildren bool) {
	switch x := n.(type) {
	case *ast.WithClause:
		for _, cte := range x.CTEs {
			c.cteNames.Add(TableName{SchemaName: c.defaultSchemaName, TableName: cte.Name.String()})
		}
	case *ast.TableName:
		var t TableName
		if x.Schema.L == "" {
			t = TableName{SchemaName: c.defaultSchemaName, TableName: x.Name.String()}
		} else {
			t = TableName{SchemaName: x.Schema.O, TableName: x.Name.String()}
		}
		if !c.cteNames.Contains(t) {
			c.tableNames.Add(t)
		}
	}
	return n, false
}

func (c *tableNameCollector) Leave(n ast.Node) (out ast.Node, ok bool) {
	return n, true
}

// CollectTableNamesFromSQL returns all referenced table names in the given Query text.
// The returned format is `schemaName.tableName`.
// TODO: handle views and CTEs.
func CollectTableNamesFromSQL(defaultSchemaName, sqlText string) (Set[TableName], error) {
	node, err := ParseOneSQL(sqlText)
	if err != nil {
		return nil, err
	}
	c := &tableNameCollector{
		defaultSchemaName: defaultSchemaName,
		tableNames:        NewSet[TableName](),
		cteNames:          NewSet[TableName]()}
	node.Accept(c)
	return c.tableNames, nil
}

// CollectTableNamesFromQueries returns all referenced table names in the given queries.
func CollectTableNamesFromQueries(queries Set[Query]) (Set[TableName], error) {
	tableNames := NewSet[TableName]()
	for _, q := range queries.ToList() {
		names, err := CollectTableNamesFromSQL(q.SchemaName, q.Text)
		if err != nil {
			return nil, err
		}
		tableNames.AddSet(names)
	}
	return tableNames, nil
}

// IsTiDBSystemTableName returns whether the given table name is a TiDB system table name.
func IsTiDBSystemTableName(t TableName) bool {
	schemaName := strings.ToLower(t.SchemaName)
	return schemaName == "information_schema" ||
		schemaName == "metrics_schema" ||
		schemaName == "performance_schema" ||
		schemaName == "mysql"
}

type selectColExtractor struct {
	selectCols       Set[Column]
	t                TableName
	underSelectField bool
}

func (e *selectColExtractor) Enter(n ast.Node) (out ast.Node, skipChildren bool) {
	switch x := n.(type) {
	case *ast.SelectField:
		e.underSelectField = true
	case *ast.ColumnNameExpr:
		if e.underSelectField {
			e.selectCols.Add(Column{
				SchemaName: e.t.SchemaName,
				TableName:  e.t.TableName,
				ColumnName: x.Name.Name.O})
		}
	}
	return n, false
}

func (e *selectColExtractor) Leave(n ast.Node) (out ast.Node, ok bool) {
	switch n.(type) {
	case *ast.SelectField:
		e.underSelectField = false
	}
	return n, true
}

// ParseSelectColumnsFromQuery returns all column names in select field.
func ParseSelectColumnsFromQuery(q Query) (Set[Column], error) {
	t, err := CollectTableNamesFromSQL(q.SchemaName, q.Text)
	if err != nil {
		return nil, err
	}
	if t.Size() != 1 { // unsupported yet
		return nil, nil
	}
	node, err := ParseOneSQL(q.Text)
	if err != nil {
		return nil, err
	}
	e := &selectColExtractor{
		selectCols: NewSet[Column](),
		t:          t.ToList()[0],
	}
	node.Accept(e)
	return e.selectCols, nil
}

type orderByColExtractor struct {
	orderByCols []Column
	t           TableName
	exit        bool
}

func (e *orderByColExtractor) Enter(n ast.Node) (out ast.Node, skipChildren bool) {
	if e.exit {
		return n, true
	}
	switch x := n.(type) {
	case *ast.OrderByClause:
		for _, byItem := range x.Items {
			colExpr, ok := byItem.Expr.(*ast.ColumnNameExpr)
			if !ok {
				e.orderByCols = nil
				e.exit = true
				return n, true
			}
			e.orderByCols = append(e.orderByCols, Column{
				SchemaName: e.t.SchemaName,
				TableName:  e.t.TableName,
				ColumnName: colExpr.Name.Name.O})
		}
	}
	return n, false
}

func (e *orderByColExtractor) Leave(n ast.Node) (out ast.Node, ok bool) {
	return n, true
}

// ParseOrderByColumnsFromQuery returns all column names in order by field.
// For a query `select ... order by c1, c2, c3`, the order by columns are `c1`, `c2` and `c3`.
func ParseOrderByColumnsFromQuery(q Query) ([]Column, error) {
	t, err := CollectTableNamesFromSQL(q.SchemaName, q.Text)
	if err != nil {
		return nil, err
	}
	if t.Size() != 1 { // unsupported yet
		return nil, nil
	}
	node, err := ParseOneSQL(q.Text)
	if err != nil {
		return nil, err
	}
	e := &orderByColExtractor{
		t: t.ToList()[0],
	}
	node.Accept(e)
	return e.orderByCols, nil
}

// ParseDNFColumnsFromQuery parses the given Query text and returns the DNF columns.
// For a query `select ... where c1=1 or c2=2 or c3=3`, the DNF columns are `c1`, `c2` and `c3`.
func ParseDNFColumnsFromQuery(q Query) (Set[Column], error) {
	t, err := CollectTableNamesFromSQL(q.SchemaName, q.Text)
	if err != nil {
		return nil, err
	}
	if t.Size() != 1 { // unsupported yet
		return nil, nil
	}
	node, err := ParseOneSQL(q.Text)
	if err != nil {
		return nil, err
	}
	e := &dnfColExtractor{
		dnfCols: NewSet[Column](),
		t:       t.ToList()[0],
	}
	node.Accept(e)
	return e.dnfCols, nil
}

type dnfColExtractor struct {
	dnfCols Set[Column]
	t       TableName
}

func (d *dnfColExtractor) Enter(n ast.Node) (node ast.Node, skipChildren bool) {
	if d.dnfCols.Size() > 0 { // already collected
		return n, true
	}
	switch x := n.(type) {
	case *ast.SelectStmt:
		cnf := flattenCNF(x.Where)
		for _, expr := range cnf {
			dnf := flattenDNF(expr)
			if len(dnf) <= 1 {
				continue
			}
			// c1=1 or c2=2 or c3=3
			var dnfCols []*ast.ColumnNameExpr
			fail := false
			for _, dnfExpr := range dnf {
				col, _ := flattenColEQConst(dnfExpr)
				if col == nil {
					fail = true
					break
				}
				dnfCols = append(dnfCols, col)
			}
			if fail {
				continue
			}
			for _, col := range dnfCols {
				d.dnfCols.Add(Column{SchemaName: d.t.SchemaName, TableName: d.t.TableName, ColumnName: col.Name.Name.O})
			}
		}
	}
	return n, false
}

func (d *dnfColExtractor) Leave(n ast.Node) (node ast.Node, ok bool) {
	return n, true
}

func flattenColEQConst(expr ast.ExprNode) (*ast.ColumnNameExpr, *driver.ValueExpr) {
	if _, ok := expr.(*ast.ParenthesesExpr); ok {
		return flattenColEQConst(expr.(*ast.ParenthesesExpr).Expr)
	}

	if op, ok := expr.(*ast.BinaryOperationExpr); ok && op.Op == opcode.EQ {
		l, r := op.L, op.R
		_, lIsCol := l.(*ast.ColumnNameExpr)
		_, lIsCon := l.(*driver.ValueExpr)
		_, rIsCol := r.(*ast.ColumnNameExpr)
		_, rIsCon := r.(*driver.ValueExpr)
		if lIsCol && rIsCon {
			return l.(*ast.ColumnNameExpr), r.(*driver.ValueExpr)
		}
		if lIsCon && rIsCol {
			return r.(*ast.ColumnNameExpr), l.(*driver.ValueExpr)
		}
	}
	return nil, nil
}

func flattenCNF(expr ast.ExprNode) []ast.ExprNode {
	if _, ok := expr.(*ast.ParenthesesExpr); ok {
		return flattenCNF(expr.(*ast.ParenthesesExpr).Expr)
	}

	var cnf []ast.ExprNode
	if op, ok := expr.(*ast.BinaryOperationExpr); ok && op.Op == opcode.LogicAnd {
		cnf = append(cnf, flattenCNF(op.L)...)
		cnf = append(cnf, flattenCNF(op.R)...)
	} else {
		cnf = append(cnf, expr)
	}
	return cnf
}

func flattenDNF(expr ast.ExprNode) []ast.ExprNode {
	if _, ok := expr.(*ast.ParenthesesExpr); ok {
		return flattenDNF(expr.(*ast.ParenthesesExpr).Expr)
	}

	var cnf []ast.ExprNode
	if op, ok := expr.(*ast.BinaryOperationExpr); ok && op.Op == opcode.LogicOr {
		cnf = append(cnf, flattenDNF(op.L)...)
		cnf = append(cnf, flattenDNF(op.R)...)
	} else {
		cnf = append(cnf, expr)
	}
	return cnf
}
