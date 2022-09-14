package fiqlparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleParse() {
	res, err := Parse("column==value")
	if err != nil {
		return
	}
	json, _ := json.Marshal(&res)
	fmt.Printf("%s", json)
	// Output:
	// {"Type":"Expr","Operator":"","Nodes":[{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"column"},{"Type":"Const","Value":"value"}]}]}
}

type testVisitor struct {
	sb strings.Builder
}

func (t *testVisitor) VisitGroupEntered()                       { t.sb.WriteString("(") }
func (t *testVisitor) VisitGroupLeft()                          { t.sb.WriteString(")") }
func (t *testVisitor) VisitOperator(operator OperatorDefintion) { t.sb.WriteString(string(operator)) }
func (t *testVisitor) VisitSelector(selector string)            { t.sb.WriteString(selector) }
func (t *testVisitor) VisitComparison(comparison ComparisonDefintion) {
	t.sb.WriteString(string(comparison))
}
func (t *testVisitor) VisitArgument(argument string) { t.sb.WriteString(argument) }

func (t *testVisitor) String() string { return t.sb.String() }

func TestParse(t *testing.T) {
	var values = []struct {
		fiql        string
		stringOuput string
		errorOutput error
	}{
		{fiql: "column=q=value", stringOuput: "column query value", errorOutput: nil},
		{fiql: "column==value", stringOuput: "column == value", errorOutput: nil},
		{fiql: "column!=value", stringOuput: "column <> value", errorOutput: nil},
		{fiql: "column=ge=value", stringOuput: "column >= value", errorOutput: nil},
		{fiql: "column=le=value", stringOuput: "column <= value", errorOutput: nil},
		{fiql: "column=gt=value", stringOuput: "column > value", errorOutput: nil},
		{fiql: "column=lt=value", stringOuput: "column < value", errorOutput: nil},
		{fiql: "column=in=value", stringOuput: "column in value", errorOutput: nil},
		{fiql: "column=q=value", stringOuput: "column query value", errorOutput: nil},
		{fiql: "column      ==        value", stringOuput: "column == value", errorOutput: nil},
		{fiql: "updated=gt=2003-12-13T00:00:00Z", stringOuput: "updated > 2003-12-13T00:00:00Z", errorOutput: nil},
		{fiql: "column==value,b==a", stringOuput: "column == value OR b == a", errorOutput: nil},
		{fiql: "column==value  ,   b==a", stringOuput: "column == value OR b == a", errorOutput: nil},
		{fiql: "     column==value  ,   b==a     ", stringOuput: "column == value OR b == a", errorOutput: nil},
		{fiql: "     column  ==  value  ,   b  ==  a     ", stringOuput: "column == value OR b == a", errorOutput: nil},
		{fiql: "(column==value)", stringOuput: "(column == value)", errorOutput: nil},
		{fiql: "column==va\\,lue,b==a", stringOuput: "column == va,lue OR b == a", errorOutput: nil},
		{fiql: "title==foo*;(updated=lt=-P1D,title==*bar)", stringOuput: "title == foo* AND (updated < -P1D OR title == *bar)", errorOutput: nil},
		{fiql: "(title==foo*);(fml==x,(xfs==a;f==fx))", stringOuput: "(title == foo*) AND (fml == x OR (xfs == a AND f == fx))", errorOutput: nil},
		{fiql: "(title==foo*,test==a,fx==fa);(fml==x)", stringOuput: "(title == foo* OR test == a OR fx == fa) AND (fml == x)", errorOutput: nil},
		{fiql: "(title==foo*);(fml==x,(xfs==a;f==fx)", stringOuput: "", errorOutput: errors.New("ln:1:36 syntax error (unclosed brace `)` )")},
		{fiql: "title=ffoo*", stringOuput: "", errorOutput: errors.New("ln:1:6 unexpected input (got `=f` but expected one of ==,!=,=gt=,=ge=,=lt=,=le=,=in=,=q=)")},
		{fiql: "title==fo,o*", stringOuput: "", errorOutput: errors.New("ln:1:12 syntax error (got `eof` but expected a value)")},

		{fiql: `a==value
		; 
		b==value`, stringOuput: "a == value AND b == value", errorOutput: nil},

		{fiql: `
		
		)`, stringOuput: "", errorOutput: errors.New("ln:3:3 syntax error (invalid closing brace `)` )")},
	}
	for _, v := range values {
		res, err := Parse(v.fiql)
		if v.errorOutput != nil {
			assert.EqualError(t, err, v.errorOutput.Error())
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, v.stringOuput, res.String())
		}

	}
}

func TestVisitor(t *testing.T) {
	p := NewParser()
	tree, err := p.Parse("(title==foo*);(fml==x,(xfs==a;f==fx))")
	assert.NoError(t, err)
	v := &testVisitor{}
	tree.Accept(v)
	assert.Equal(t, v.String(), "(title==foo*)AND(fml==xOR(xfs==aANDf==fx))")
}

func TestJsonMarshall(t *testing.T) {
	var values = []struct {
		fiql        string
		stringOuput string
		errorOutput error
	}{
		{fiql: "column=q=value", stringOuput: `{"Type":"Expr","Operator":"","Nodes":[{"Type":"Binary","Operator":"query","Nodes":[{"Type":"Const","Value":"column"},{"Type":"Const","Value":"value"}]}]}`, errorOutput: nil},
		{fiql: "column==value", stringOuput: `{"Type":"Expr","Operator":"","Nodes":[{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"column"},{"Type":"Const","Value":"value"}]}]}`, errorOutput: nil},
		{fiql: "column!=value", stringOuput: `{"Type":"Expr","Operator":"","Nodes":[{"Type":"Binary","Operator":"\u003c\u003e","Nodes":[{"Type":"Const","Value":"column"},{"Type":"Const","Value":"value"}]}]}`, errorOutput: nil},
		{fiql: "title==foo*;(updated=lt=-P1D,title==*bar)", stringOuput: `{"Type":"Expr","Operator":"","Nodes":[{"Type":"Binary","Operator":"AND","Nodes":[{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"title"},{"Type":"Const","Value":"foo*"}]},{"Type":"Group","Nodes":[{"Type":"Binary","Operator":"OR","Nodes":[{"Type":"Binary","Operator":"\u003c","Nodes":[{"Type":"Const","Value":"updated"},{"Type":"Const","Value":"-P1D"}]},{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"title"},{"Type":"Const","Value":"*bar"}]}]}]}]}]}`, errorOutput: nil},
		{fiql: "(title==foo*);(fml==x,(xfs==a;f==fx))", stringOuput: `{"Type":"Expr","Operator":"","Nodes":[{"Type":"Binary","Operator":"AND","Nodes":[{"Type":"Group","Nodes":[{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"title"},{"Type":"Const","Value":"foo*"}]}]},{"Type":"Group","Nodes":[{"Type":"Binary","Operator":"OR","Nodes":[{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"fml"},{"Type":"Const","Value":"x"}]},{"Type":"Group","Nodes":[{"Type":"Binary","Operator":"AND","Nodes":[{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"xfs"},{"Type":"Const","Value":"a"}]},{"Type":"Binary","Operator":"==","Nodes":[{"Type":"Const","Value":"f"},{"Type":"Const","Value":"fx"}]}]}]}]}]}]}]}`, errorOutput: nil},
	}

	for _, v := range values {
		res, err := Parse(v.fiql)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		j, err := json.Marshal(&res)
		assert.NoError(t, err)
		assert.Equal(t, v.stringOuput, string(j))
	}
}
