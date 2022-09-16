package fiqlparser

import (
	"fmt"
	"strings"
)

// SimpleTestVisitor is a simple visitor used to collect
// all visited data in a string, it adheres to the visitor interface
type SimpleTestVisitor struct {
	sb strings.Builder
}

// VisitExpressionEntered is called when a expression is entered
func (t *SimpleTestVisitor) VisitExpressionEntered() { t.sb.WriteString("(") }

// VisitExpressionLeft is called when a expression is left
func (t *SimpleTestVisitor) VisitExpressionLeft() { t.sb.WriteString(")") }

// VisitOperator is called when a operator is visited
func (t *SimpleTestVisitor) VisitOperator(operator OperatorDefintion) {
	t.sb.WriteRune(' ')
	t.sb.WriteString(string(operator))
	t.sb.WriteByte(' ')
}

// VisitSelector is called when a selector is visited
func (t *SimpleTestVisitor) VisitSelector(selector string) { t.sb.WriteString(selector) }

// VisitComparison is called when a comparison is visited
func (t *SimpleTestVisitor) VisitComparison(comparison ComparisonDefintion) {
	t.sb.WriteString(string(comparison))
}

// VisitArgument is called when an argument is visited
func (t *SimpleTestVisitor) VisitArgument(argument string, valueCtx ValueContext) {
	if valueCtx.StartsWithWildcard() {
		t.sb.WriteString("*")
	}
	t.sb.WriteString(argument)
	if valueCtx.EndsWithWildcard() {
		t.sb.WriteString("*")
	}
}

// String returns the colected data as string
func (t *SimpleTestVisitor) String() string { return t.sb.String() }

// ExampleVisitor demonstrates how to use a simple visitor on a expression
func Example() {
	p := NewParser()
	tree, err := p.Parse("(title==foo*);(fml==x,(xfs==a;f==fx))")
	if err != nil {
		return
	}
	v := &SimpleTestVisitor{}
	tree.Accept(v)
	fmt.Print(v.String())
	// Output:
	// ((title==foo*) AND (fml==x OR (xfs==a AND f==fx)))
}
