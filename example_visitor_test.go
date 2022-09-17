// This example demonstrates a basic parser run
// with a simple visitor
package fiqlparser_test

import (
	"fmt"
	"strings"

	fq "github.com/eisenwinter/fiql-parser"
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
func (t *SimpleTestVisitor) VisitOperator(operatorCtx fq.OperatorContext) {
	t.sb.WriteRune(' ')
	t.sb.WriteString(string(operatorCtx.Operator()))
	t.sb.WriteByte(' ')
}

// VisitSelector is called when a selector is visited
func (t *SimpleTestVisitor) VisitSelector(selectorCtx fq.SelectorContext) {
	t.sb.WriteString(selectorCtx.Selector())
}

// VisitComparison is called when a comparison is visited
func (t *SimpleTestVisitor) VisitComparison(comparisonCtx fq.ComparisonContext) {
	t.sb.WriteString(string(comparisonCtx.Comparison()))
}

// VisitArgument is called when an argument is visited
func (t *SimpleTestVisitor) VisitArgument(argumentCtx fq.ArgumentContext) {
	if argumentCtx.StartsWithWildcard() {
		t.sb.WriteString("*")
	}
	t.sb.WriteString(argumentCtx.AsString())
	if argumentCtx.EndsWithWildcard() {
		t.sb.WriteString("*")
	}
}

// String returns the colected data as string
func (t *SimpleTestVisitor) String() string { return t.sb.String() }

// ExampleVisitor demonstrates how to use a simple visitor on a expression
func Example() {
	p := fq.NewParser()
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
