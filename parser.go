package fiqlparser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// NodeType defines the type of node in the ast
type NodeType string

// NodeTypeExpression is the root element of any expression
const NodeTypeExpression NodeType = "Expr"

// NodeTypeBinary is a binary expression
const NodeTypeBinary NodeType = "Binary"

// NodeTypeConstant is a constant value expression
const NodeTypeConstant NodeType = "Const"

// NodeTypeGroup is a expression used to group items
const NodeTypeGroup NodeType = "Group"

// WalkOperation defines its coming or going, mainly used for braces
type WalkOperation bool

// WalkEntered coming
const WalkEntered WalkOperation = true

// WalkLeave going
const WalkLeave WalkOperation = false

// OperatorDefintion defines the two operators fiql has
type OperatorDefintion string

// OperatorOR defines the OR operation
const OperatorOR OperatorDefintion = "OR"

// OperatorAND defines the AND operation
const OperatorAND OperatorDefintion = "AND"

// ComparisonDefintion defines the fiql + custom comparisons
type ComparisonDefintion string

// ComparisonEq equal comparison
const ComparisonEq ComparisonDefintion = "=="

// ComparisonNeq not equal comparison
const ComparisonNeq ComparisonDefintion = "<>"

// ComparisonEq greater comparison
const ComparisonGt ComparisonDefintion = ">"

// ComparisonEq less comparison
const ComparisonLt ComparisonDefintion = "<"

// ComparisonEq greater or equal comparison
const ComparisonGte ComparisonDefintion = ">="

// ComparisonEq less or equal comparison
const ComparisonLte ComparisonDefintion = "<="

// ComparisonEq in comparison
const ComparisonIn ComparisonDefintion = "in"

// ComparisonEq query comparison
const ComparisonQ ComparisonDefintion = "query"

//Basically follow naming of https://datatracker.ietf.org/doc/html/draft-nottingham-atompub-fiql-00#section-3.2

// NodeVisitor is used to visit the tree
type NodeVisitor interface {
	// VisitGroupEntered is called when a group is entered
	VisitGroupEntered()

	// VisitGroupEntered is called when a group is left
	VisitGroupLeft()

	// VisitOperator is called when a operator is visited
	VisitOperator(operator OperatorDefintion)

	// VisitSelector is called when a selector is visited
	VisitSelector(selector string)

	// VisitComparison is called when a comparison is visited
	VisitComparison(comparison ComparisonDefintion)

	// VisitArgument is called when a argument is visited
	VisitArgument(argument string)
}

// Node represents a AST node
type Node interface {
	// NodeType - node type in the AST - the root node will always be expression
	NodeType() NodeType
	// Add adds a child node to the node
	Add(node Node)
	// String prints the node
	String() string
	// Walk allows to iterate over the node and its child nodes
	Walk(fx func(Node, WalkOperation))
	// Returns the children of this node
	Children() []Node

	// Accepts a Visitor
	Accept(visitor NodeVisitor)
}

// Expression is the root node
type Expression struct {
	nodes []Node
}

// NodeType NodeTypeExpression
func (e *Expression) NodeType() NodeType {
	return NodeTypeExpression
}

// Add adds a childnode
func (e *Expression) Add(node Node) {
	e.nodes = append(e.nodes, node)
}

// Walk walks the expression
func (e *Expression) Walk(fx func(Node, WalkOperation)) {
	for _, v := range e.nodes {
		v.Walk(fx)
	}
	fx(e, WalkLeave)
}

// Accept accepts a vistor to visit the tree
func (e *Expression) Accept(visitor NodeVisitor) {
	for _, v := range e.nodes {
		v.Accept(visitor)
	}
}

// MarshalJSON overloading for json marshalling
func (e *Expression) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Type     string
		Operator string
		Nodes    []Node
	}{
		Type:  string(e.NodeType()),
		Nodes: e.nodes,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (e *Expression) String() string {
	var b strings.Builder
	e.Walk(func(n Node, w WalkOperation) {
		switch n.NodeType() {
		case NodeTypeExpression:
			break
		case NodeTypeGroup:
			if w == WalkEntered {
				b.WriteString("(")
			} else {
				b.WriteString(")")
			}
		default:
			b.WriteString(n.String())
		}
	})
	return b.String()
}

func (e *Expression) Children() []Node {
	return e.nodes
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
func (e *binaryExpression) Walk(fx func(Node, WalkOperation)) {
	if e.nodes[0] != nil {
		e.nodes[0].Walk(fx)
	}
	fx(e, WalkEntered)
	if e.nodes[1] != nil {
		e.nodes[1].Walk(fx)
	}

}

// Accept accepts a vistor to visit the tree
func (e *binaryExpression) Accept(visitor NodeVisitor) {
	if e.nodes[0] != nil {
		e.nodes[0].Accept(visitor)
	}
	//conjs
	if e.operator == "AND" || e.operator == "OR" {
		visitor.VisitOperator(OperatorDefintion(e.operator))
	} else {
		visitor.VisitComparison(ComparisonDefintion(e.operator))
	}
	if e.nodes[1] != nil {
		e.nodes[1].Accept(visitor)
	}
}

func (e *binaryExpression) Children() []Node {
	return []Node{e.nodes[0], e.nodes[1]}
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
	return " " + e.operator + " "
}

type constantExpression struct {
	selector bool
	value    string
}

func (e *constantExpression) NodeType() NodeType {
	return NodeTypeConstant
}

func (e *constantExpression) Add(node Node) {
	panic("constant should have a child")
}

func (e *constantExpression) Walk(fx func(Node, WalkOperation)) {
	fx(e, WalkEntered)
}

func (e *constantExpression) Accept(visitor NodeVisitor) {
	if e.selector {
		visitor.VisitSelector(e.value)
	} else {
		visitor.VisitArgument(e.value)
	}

}

func (e *constantExpression) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Type  string
		Value string
	}{
		Type:  string(e.NodeType()),
		Value: e.value,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (e *constantExpression) String() string {
	return e.value
}

func (e *constantExpression) Children() []Node {
	return []Node{}
}

type groupExpression struct {
	nodes []Node
}

func (e groupExpression) NodeType() NodeType {
	return NodeTypeGroup
}

func (e *groupExpression) Add(node Node) {
	e.nodes = append(e.nodes, node)
}

func (e *groupExpression) Walk(fx func(Node, WalkOperation)) {
	fx(e, WalkEntered)
	for _, v := range e.nodes {
		v.Walk(fx)
	}
	fx(e, WalkLeave)
}

// Accept accepts a vistor to visit the tree
func (e *groupExpression) Accept(visitor NodeVisitor) {
	visitor.VisitGroupEntered()
	for _, v := range e.nodes {
		v.Accept(visitor)
	}
	visitor.VisitGroupLeft()
}

func (e *groupExpression) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Type  string
		Nodes []Node
	}{
		Type:  string(e.NodeType()),
		Nodes: e.nodes,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (e *groupExpression) String() string {
	return "group"
}

func (e *groupExpression) Children() []Node {
	return e.nodes
}

// Parser is the fiql parser
type Parser struct {
	lex *lexer
}

func (p *Parser) handleGroup(parent Node) (Node, error) {
	group := &groupExpression{nodes: make([]Node, 0)}
	n, err := p.build(group)
	if err != nil {
		return group, err
	}
	group.Add(n)
	return group, nil
}

func (p *Parser) handleBinaryExpression(t tokenType) (Node, error) {
	bin := &binaryExpression{nodes: [2]Node{nil, nil}}
	bin.operator = t.String()
	bin.Add(&constantExpression{value: p.lex.lastValue(), selector: true})
	t, err := p.lex.ConsumeToken()
	if err != nil {
		return bin, err
	}
	if isCompareToken(t) {
		bin.operator = t.String()
	} else {
		return bin, fmt.Errorf("ln:%d:%d syntax error (got `%s` but expected a value)", p.lex.ln, p.lex.posInLine, t.String())
	}

	t, err = p.lex.ConsumeToken()
	if err != nil {
		return bin, err
	}
	if t == tokenValue {
		bin.Add(&constantExpression{value: p.lex.lastValue()})
	} else {
		return bin, fmt.Errorf("ln:%d:%d syntax error (got `%s` but expected a value)", p.lex.ln, p.lex.posInLine, t.String())
	}

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
	return bin, nil
}

func (p *Parser) build(parent Node) (Node, error) {
	t, err := p.lex.ConsumeToken()
	if err != nil {
		return parent, err
	}
	if t == tokenEOF {
		return parent, nil
	}
	if t == tokenBraceClose {
		return parent, fmt.Errorf("ln:%d:%d syntax error (invalid closing brace `)` )", p.lex.ln, p.lex.posInLine)
	}

	if t == tokenBraceOpen {
		group, err := p.handleGroup(parent)
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
			t, err = p.lex.ConsumeToken()
			if err != nil {
				return parent, err
			}
			conj := &binaryExpression{nodes: [2]Node{nil, nil}}
			conj.operator = t.String()
			conj.Add(group)

			rhs, err := p.build(conj)
			if err != nil {
				return conj, err
			}
			conj.Add(rhs)

			parent.Add(conj)
			return parent, nil
		}
		if parent.NodeType() == NodeTypeExpression {
			parent.Add(group)
			return parent, nil
		}
		return group, nil

	}

	if t == tokenValue {
		//TODO: consider unary (its in the draft but i dont see how its valuable for my purpose)
		binary, err := p.handleBinaryExpression(t)
		if parent.NodeType() == NodeTypeExpression {
			parent.Add(binary)
			return parent, err
		}
		return binary, err
	}
	return parent, err

}

func (p *Parser) Parse(input string) (Expression, error) {
	p.lex = &lexer{input, 0, 1, 0, ""}
	exp := Expression{}
	_, err := p.build(&exp)
	return exp, err
}

// NewParser returns a new figl parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse instant parses the supplied figl
func Parse(input string) (Expression, error) {
	p := &Parser{}
	return p.Parse(input)
}
