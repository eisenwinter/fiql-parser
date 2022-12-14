// Package fiqlparser provides a simple fiql parser.
//
// The parser does not adhere 100% to the fiql spec
// which can be found https://datatracker.ietf.org/doc/html/draft-nottingham-atompub-fiql-00.
//
// The parser produces a walkable AST which can be walked by using a visitor.
package fiqlparser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NodeType defines the type of node in the ast
type NodeType string

// NodeTypeExpression is the root element of any expression
const NodeTypeExpression NodeType = "Expr"

// NodeTypeBinary is a binary expression
const NodeTypeBinary NodeType = "Binary"

// NodeTypeConstant is a constant value expression
const NodeTypeConstant NodeType = "Const"

// OperatorDefintion defines the two operators fiql has
type OperatorDefintion string

// OperatorOR defines the OR operation
// Associativity: Left to right
const OperatorOR OperatorDefintion = "OR"

// OperatorAND defines the AND operation
// Associativity: Left to right
const OperatorAND OperatorDefintion = "AND"

// ComparisonDefintion defines the fiql + custom comparisons
type ComparisonDefintion string

// ComparisonEq equal comparison
const ComparisonEq ComparisonDefintion = "=="

// ComparisonNeq not equal comparison
const ComparisonNeq ComparisonDefintion = "<>"

// ComparisonGt greater comparison
const ComparisonGt ComparisonDefintion = ">"

// ComparisonLt less comparison
const ComparisonLt ComparisonDefintion = "<"

// ComparisonGte greater or equal comparison
const ComparisonGte ComparisonDefintion = ">="

// ComparisonLte less or equal comparison
const ComparisonLte ComparisonDefintion = "<="

// ValueRecommendation suggests a detected datatype for a attribute
type ValueRecommendation string

// ValueRecommendationString suggests a string attribute
const ValueRecommendationString ValueRecommendation = "string"

// ValueRecommendationDateTime suggests a date attribute
const ValueRecommendationDateTime ValueRecommendation = "datetime"

// ValueRecommendationDuration suggests a duration attribute
const ValueRecommendationDuration ValueRecommendation = "duration"

// ValueRecommendationNumber suggests a number attribute
const ValueRecommendationNumber ValueRecommendation = "number"

// ArgumentContext habours the value and
// supplies the recommended type + conversion helpers
type ArgumentContext struct {
	pre  bool
	post bool
	r    ValueRecommendation
	val  string
}

// ValueRecommendation returns the value recommendation
func (c ArgumentContext) ValueRecommendation() ValueRecommendation {
	return c.r
}

// StartsWithWildcard indicates whether or not the given argument starts with a wildcard
func (c ArgumentContext) StartsWithWildcard() bool {
	return c.pre
}

// EndsWithWildcard indicates whether or not the given argument ends with a wildcard
func (c ArgumentContext) EndsWithWildcard() bool {
	return c.post
}

// AsString returns the argument as string
func (c ArgumentContext) AsString() string {
	return c.val
}

// AsDuration is a helper method for converting duration values
func (c ArgumentContext) AsDuration() (ISO8601Duration, error) {
	return durationConverter.tryParseISO8601Duration(c.val)
}

// AsTime is a helper method for converting duration values
func (c ArgumentContext) AsTime() (time.Time, error) {
	return time.Parse(time.RFC3339, c.val)
}

// AsInt returns the underlying value as int
func (c ArgumentContext) AsInt() (int, error) {
	return strconv.Atoi(c.val)
}

// AsFloat64 returns the underlying value as float64
func (c ArgumentContext) AsFloat64() (float64, error) {
	return strconv.ParseFloat(c.val, 64)
}

// SelectorContext contains the selector details
type SelectorContext struct {
	unary    bool
	selector string
}

// Selector returns the selector as string
func (s SelectorContext) Selector() string {
	return s.selector
}

// IsUnary returns true if the selector has no constraint
func (s SelectorContext) IsUnary() bool {
	return s.unary
}

// OperatorContext contains the operator details
type OperatorContext struct {
	op OperatorDefintion
}

// Operator returns the operator
func (c OperatorContext) Operator() OperatorDefintion {
	return c.op
}

// ComparisonContext contains the comerator details
type ComparisonContext struct {
	comparison ComparisonDefintion
}

// Comparison returns the used comparison
func (c ComparisonContext) Comparison() ComparisonDefintion {
	return c.comparison
}

//Basically follow naming of https://datatracker.ietf.org/doc/html/draft-nottingham-atompub-fiql-00#section-3.2

// NodeVisitor is used to visit the tree
type NodeVisitor interface {
	// VisitExpressionEntered is called when a expression is entered
	VisitExpressionEntered()

	// VisitExpressionLeft is called when a expression is left
	VisitExpressionLeft()

	// VisitOperator is called when a operator is visited
	VisitOperator(operatorCtx OperatorContext)

	// VisitSelector is called when a selector is visited
	VisitSelector(selectorCtx SelectorContext)

	// VisitComparison is called when a comparison is visited
	VisitComparison(comparisonCtx ComparisonContext)

	// VisitArgument is called when a argument is visited
	VisitArgument(argumentCtx ArgumentContext)
}

// Node represents a AST node
type Node interface {
	// NodeType - node type in the AST - the root node will always be expression
	NodeType() NodeType
	// String prints the node
	String() string
	// Returns the children of this node
	Children() []Node
	// Accepts a Visitor
	Accept(visitor NodeVisitor)
	// Add adds a child node to this node
	Add(Node)

	// isRoot indicates the root node
	isRoot() bool
}

// Expression is the root node
type Expression struct {
	node Node
	root bool
}

func (e *Expression) isRoot() bool {
	return e.root
}

// NodeType NodeTypeExpression
func (e *Expression) NodeType() NodeType {
	return NodeTypeExpression
}

// Accept accepts a vistor to visit the tree
func (e *Expression) Accept(visitor NodeVisitor) {
	visitor.VisitExpressionEntered()
	if e.node != nil {
		e.node.Accept(visitor)
	}
	visitor.VisitExpressionLeft()
}

// Add adds a child to the node, it will panic if more than one child exists on a expression node
func (e *Expression) Add(node Node) {
	if e.node != nil {
		panic("node may not have more than one child")
	}
	e.node = node
}

// MarshalJSON overloading for json marshalling
func (e *Expression) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Type     string
		Operator string
		Nodes    []Node
	}{
		Type:  string(e.NodeType()),
		Nodes: []Node{e.node},
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (e *Expression) String() string {
	var b strings.Builder
	b.WriteRune('(')
	for _, v := range e.Children() {
		b.WriteString(v.String())
	}
	b.WriteRune(')')
	return b.String()
}

// Children returns the children of this expression
func (e *Expression) Children() []Node {
	return []Node{e.node}
}

type binaryExpression struct {
	operator string
	nodes    [2]Node
}

func (e *binaryExpression) NodeType() NodeType {
	return NodeTypeBinary
}

func (e *binaryExpression) Add(node Node) {
	if e.nodes[0] == nil {
		e.nodes[0] = node
		return
	}
	if e.nodes[1] == nil {
		e.nodes[1] = node
		return
	}
	panic("binary node cant hold more than two values")
}

// Accept accepts a vistor to visit the tree
func (e *binaryExpression) Accept(visitor NodeVisitor) {
	if e.nodes[0] != nil {
		e.nodes[0].Accept(visitor)
	}
	//conjs
	if e.operator == "AND" || e.operator == "OR" {
		visitor.VisitOperator(OperatorContext{op: OperatorDefintion(e.operator)})
	} else {
		visitor.VisitComparison(ComparisonContext{comparison: ComparisonDefintion(e.operator)})
	}
	if e.nodes[1] != nil {
		e.nodes[1].Accept(visitor)
	}
}

func (e *binaryExpression) Children() []Node {
	nodes := make([]Node, 0)
	if e.nodes[0] != nil {
		nodes = append(nodes, e.nodes[0])
	}
	if e.nodes[1] != nil {
		nodes = append(nodes, e.nodes[1])
	}
	return nodes
}

func (e *binaryExpression) isRoot() bool {
	return false
}

func (e *binaryExpression) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Type     string
		Operator string
		Nodes    [2]Node
	}{
		Type:     string(e.NodeType()),
		Operator: e.operator,
		Nodes:    e.nodes,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (e *binaryExpression) String() string {
	var b strings.Builder
	if e.nodes[0] != nil {
		b.WriteString(e.nodes[0].String())
	}
	b.WriteRune(' ')
	b.WriteString(e.operator)
	b.WriteRune(' ')
	if e.nodes[1] != nil {
		b.WriteString(e.nodes[1].String())
	}

	return b.String()
}

type constantExpression struct {
	prefixWildcard bool
	suffixWildcard bool
	selector       bool
	value          string
	recommended    ValueRecommendation
	unary          bool
}

func (e *constantExpression) isRoot() bool {
	return false
}

func (e *constantExpression) NodeType() NodeType {
	return NodeTypeConstant
}

func (e *constantExpression) Add(node Node) {
	panic("constant should not have a child")
}

func (e *constantExpression) Accept(visitor NodeVisitor) {
	if e.selector {
		visitor.VisitSelector(SelectorContext{unary: e.unary, selector: e.value})
	} else {
		visitor.VisitArgument(ArgumentContext{
			pre:  e.prefixWildcard,
			post: e.suffixWildcard,
			r:    e.recommended,
			val:  e.value,
		})
	}

}

func (e *constantExpression) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Type  string
		Value string
	}{
		Type:  string(e.NodeType()),
		Value: e.String(),
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (e *constantExpression) String() string {
	var b strings.Builder
	if e.prefixWildcard {
		b.WriteRune('*')
	}
	b.WriteString(e.value)
	if e.suffixWildcard {
		b.WriteRune('*')
	}
	return b.String()
}

func (e *constantExpression) Children() []Node {
	return []Node{}
}

// Parser is the fiql parser
type Parser struct {
	lex *lexer
}

func (p *Parser) handleSubExpression(parent Node) (Node, error) {
	expr := &Expression{node: nil}
	n, err := p.build(expr)
	if err != nil {
		return expr, err
	}
	expr.node = n
	return expr, nil
}

var numericRegex = regexp.MustCompile(`^(\+|-|)[0-9\.]+$`)
var durationRegex = regexp.MustCompile(`^(\+|-|)P(?:\d+(?:\.\d+)?Y)?(?:\d+(?:\.\d+)?M)?(?:\d+(?:\.\d+)?W)?(?:\d+(?:\.\d+)?D)?(?:T(?:\d+(?:\.\d+)?H)?(?:\d+(?:\.\d+)?M)?(?:\d+(?:\.\d+)?S)?)?$`)

func isDateValue(stringDate string) bool {
	_, err := time.Parse(time.RFC3339, stringDate)
	return err == nil
}

type argumentValidator func(string) (bool, ValueRecommendation, string)

func numberOrDateExpressionValidator(i string) (bool, ValueRecommendation, string) {
	if numericRegex.MatchString(i) {
		return true, ValueRecommendationNumber, ""
	}
	//time or duration e.g. 2003-12-13T18:30:02Z or  -P1D12
	if isDateValue(i) {
		return true, ValueRecommendationDateTime, ""
	}
	if durationRegex.MatchString(i) {
		return true, ValueRecommendationDuration, ""
	}

	return false, ValueRecommendationString, "number or date or duration"
}

func defaultValidator(i string) (bool, ValueRecommendation, string) {
	if isDateValue(i) {
		return true, ValueRecommendationDateTime, ""
	}
	if durationRegex.MatchString(i) {
		return true, ValueRecommendationDuration, ""
	}
	if numericRegex.MatchString(i) {
		return true, ValueRecommendationNumber, ""
	}
	return true, ValueRecommendationString, ""
}

func (p *Parser) handleArgumentConstant(validator argumentValidator) (Node, error) {
	t, err := p.lex.ConsumeToken()
	if err != nil {
		return nil, err
	}
	prefixWildcard := false
	if t == tokenWildcard {
		t, err = p.lex.ConsumeToken()
		if err != nil {
			return nil, err
		}
		prefixWildcard = true
	}
	if t == tokenValue {
		ok, rec, msg := validator(p.lex.lastValue())
		if !ok {
			return nil, fmt.Errorf("ln:%d:%d syntax error (got `%s` but expected %s)", p.lex.ln, p.lex.posInLine, p.lex.lastValue(), msg)
		}
		con := &constantExpression{prefixWildcard: prefixWildcard, value: p.lex.lastValue(), recommended: rec}
		n, _, err := p.lex.PeekNextToken()
		if err != nil {
			return nil, err
		}
		if n == tokenWildcard {
			_, err = p.lex.ConsumeToken()
			if err != nil {
				return nil, err
			}
			con.suffixWildcard = true
		}
		return con, nil
	}
	return nil, fmt.Errorf("ln:%d:%d syntax error (got `%s` but expected a value)", p.lex.ln, p.lex.posInLine, t.String())
}

func (p Parser) handleUnaryExpression(parent Node) (Node, error) {
	unary := &constantExpression{value: p.lex.lastValue(), selector: true, recommended: ValueRecommendationString, unary: true}
	next, _, err := p.lex.PeekNextToken()
	if err != nil {
		return unary, err
	}
	if isLogicToken(next) {
		t, err := p.lex.ConsumeToken()
		if err != nil {
			return unary, err
		}
		conj := &binaryExpression{nodes: [2]Node{nil, nil}}
		conj.operator = t.String()
		conj.Add(unary)
		rhs, err := p.build(conj)
		if err != nil {
			return conj, err
		}
		conj.Add(rhs)
		return conj, nil
	}
	if isCompareToken(next) {
		return unary, fmt.Errorf("ln:%d:%d dangling comparator", p.lex.ln, p.lex.posInLine)
	}
	if next == tokenBraceClose && parent.isRoot() {
		return unary, fmt.Errorf("ln:%d:%d syntax error (invalid closing brace `)` )", p.lex.ln, p.lex.posInLine)
	}
	return unary, nil
}

func (p *Parser) handleBinaryExpression(t tokenType, parent Node) (Node, error) {
	bin := &binaryExpression{nodes: [2]Node{nil, nil}}
	bin.operator = t.String()
	bin.Add(&constantExpression{value: p.lex.lastValue(), selector: true, recommended: ValueRecommendationString})
	t, err := p.lex.ConsumeToken()
	if err != nil {
		return bin, err
	}
	if isCompareToken(t) {
		bin.operator = t.String()
	} else {
		return bin, fmt.Errorf("ln:%d:%d syntax error (got `%s` but expected a value)", p.lex.ln, p.lex.posInLine, t.String())
	}

	validator := defaultValidator
	if isNumberOrDateComparision(t) {
		validator = numberOrDateExpressionValidator
	}
	con, err := p.handleArgumentConstant(validator)
	if err != nil {
		return bin, err
	}
	bin.Add(con)

	next, _, err := p.lex.PeekNextToken()
	if err != nil {
		return bin, err
	}
	if isLogicToken(next) {
		t, err = p.lex.ConsumeToken()
		if err != nil {
			return bin, err
		}
		conj := &binaryExpression{nodes: [2]Node{nil, nil}}
		conj.operator = t.String()
		conj.Add(bin)
		rhs, err := p.build(conj)
		if err != nil {
			return conj, err
		}
		conj.Add(rhs)
		return conj, nil
	}
	if isCompareToken(next) {
		return bin, fmt.Errorf("ln:%d:%d dangling comparator", p.lex.ln, p.lex.posInLine)
	}
	if next == tokenBraceClose && parent.isRoot() {
		return bin, fmt.Errorf("ln:%d:%d syntax error (invalid closing brace `)` )", p.lex.ln, p.lex.posInLine)
	}
	return bin, nil
}

func (p *Parser) checkDanglingChild(n Node) bool {
	if n.NodeType() == NodeTypeBinary {
		if len(n.Children()) != 2 {
			return true
		}
	}
	return false
}

// checkImpossibleTokensOnEnter checks for tokens that should not appear on enter `build`
func (p *Parser) checkImpossibleTokensOnEnter(t tokenType) error {
	if t == tokenBraceClose {
		return fmt.Errorf("ln:%d:%d syntax error (invalid closing brace `)` )", p.lex.ln, p.lex.posInLine)
	}

	if isLogicToken(t) {
		return fmt.Errorf("ln:%d:%d dangling operator", p.lex.ln, p.lex.posInLine)
	}

	if isCompareToken(t) {
		return fmt.Errorf("ln:%d:%d dangling comparator", p.lex.ln, p.lex.posInLine)
	}
	return nil
}

// checkForEOF checks for correct end of file
func (p *Parser) checkForEOF(t tokenType, node Node) (bool, error) {
	if t == tokenEOF {
		if p.checkDanglingChild(node) {
			return true, fmt.Errorf("ln:%d:%d dangling operator", p.lex.ln, p.lex.posInLine)

		}
		return true, nil
	}
	return false, nil
}

func (p *Parser) checkEndOrError(t tokenType, parent Node) (bool, error) {
	if ok, err := p.checkForEOF(t, parent); ok {
		return ok, err
	}
	if err := p.checkImpossibleTokensOnEnter(t); err != nil {
		return true, err
	}
	return false, nil
}

func (p *Parser) mergeSubExpression(sub, parent Node) (Node, error) {
	t, err := p.lex.ConsumeToken()
	if err != nil {
		return parent, err
	}
	conj := &binaryExpression{nodes: [2]Node{nil, nil}}
	conj.operator = t.String()
	conj.Add(sub)

	rhs, err := p.build(conj)
	if err != nil {
		return conj, err
	}
	conj.Add(rhs)
	parent.Add(conj)
	return parent, nil
}

func (p *Parser) build(parent Node) (Node, error) {
	t, err := p.lex.ConsumeToken()
	if err != nil {
		return parent, err
	}
	if ok, err := p.checkEndOrError(t, parent); ok {
		return parent, err
	}
	if t == tokenBraceOpen {
		sub, err := p.handleSubExpression(parent)
		if err != nil {
			return parent, err
		}
		t, err := p.lex.ConsumeToken()
		if err != nil {
			return parent, err
		}
		if t != tokenBraceClose {
			return parent, fmt.Errorf("ln:%d:%d syntax error (unclosed brace `)` )", p.lex.ln, p.lex.posInLine)
		}

		next, _, err := p.lex.PeekNextToken()
		if err != nil {
			return parent, err
		}
		if isLogicToken(next) {
			return p.mergeSubExpression(sub, parent)
		}
		if parent.NodeType() == NodeTypeExpression {
			parent.Add(sub)
			return parent, nil
		}
		return sub, nil
	}

	if t == tokenValue {
		next, _, err := p.lex.PeekNextToken()
		if err != nil {
			return parent, err
		}
		var nextExpr Node
		if isSeperatorUnary(next) {
			nextExpr, err = p.handleUnaryExpression(parent)
		} else {
			nextExpr, err = p.handleBinaryExpression(t, parent)
		}
		if parent.isRoot() {
			parent.Add(nextExpr)
			return parent, err
		}
		return nextExpr, err
	}
	return parent, err

}

// Parse parses the supplied fiql and returns either a Expression or an error
func (p *Parser) Parse(input string) (Expression, error) {
	p.lex = &lexer{[]rune(input), 0, 1, 0, ""}
	exp := Expression{root: true}
	_, err := p.build(&exp)
	return exp, err
}

// NewParser returns a new fiql parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse instant parses the supplied fiql and returns either a Expression or an error
func Parse(input string) (Expression, error) {
	p := &Parser{}
	return p.Parse(input)
}
