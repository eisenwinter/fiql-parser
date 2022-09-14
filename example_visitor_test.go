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

// VisitGroupEntered is called when a group is entered
func (t *SimpleTestVisitor) VisitGroupEntered() { t.sb.WriteString("(") }

// VisitGroupLeft is called when a group is left
func (t *SimpleTestVisitor) VisitGroupLeft() { t.sb.WriteString(")") }

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
func (t *SimpleTestVisitor) VisitArgument(argument string) { t.sb.WriteString(argument) }

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
	// (title==foo*) AND (fml==x OR (xfs==a AND f==fx))
}
