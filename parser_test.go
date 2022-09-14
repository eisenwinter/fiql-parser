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
func (t *testVisitor) VisitArgument(argument string, recommendedValueType ValueRecommendation) {
	t.sb.WriteString(argument)
}

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
		{fiql: "column=ge=1", stringOuput: "column >= 1", errorOutput: nil},
		{fiql: "column=le=1", stringOuput: "column <= 1", errorOutput: nil},
		{fiql: "column=gt=1.1", stringOuput: "column > 1.1", errorOutput: nil},
		{fiql: "column=lt=100", stringOuput: "column < 100", errorOutput: nil},
		{fiql: "column=lt=+100", stringOuput: "column < +100", errorOutput: nil},
		{fiql: "column=lt=-100", stringOuput: "column < -100", errorOutput: nil},
		{fiql: "column=in=[a]", stringOuput: "column in [a]", errorOutput: nil},
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
		{fiql: "column=ge=invalid", stringOuput: "", errorOutput: errors.New("ln:1:17 syntax error (got `invalid` but expected number or date or duration)")},
		{fiql: "column=in=123", stringOuput: "", errorOutput: errors.New("ln:1:13 syntax error (got `123` but expected in clause [a*b*c])")},
		{fiql: "column=le=P1.4Y2M", stringOuput: "column <= P1.4Y2M", errorOutput: nil},
		{fiql: "column=le=+P5W", stringOuput: "column <= +P5W", errorOutput: nil},
		{fiql: "column=le=-P5W", stringOuput: "column <= -P5W", errorOutput: nil},
		{fiql: "column=lt=P3DT4H59M", stringOuput: "column < P3DT4H59M", errorOutput: nil},
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

type testTypeVisitor struct {
	sb strings.Builder
}

func (t *testTypeVisitor) VisitGroupEntered()                             {}
func (t *testTypeVisitor) VisitGroupLeft()                                {}
func (t *testTypeVisitor) VisitOperator(operator OperatorDefintion)       {}
func (t *testTypeVisitor) VisitSelector(selector string)                  {}
func (t *testTypeVisitor) VisitComparison(comparison ComparisonDefintion) {}
func (t *testTypeVisitor) VisitArgument(argument string, recommendedValueType ValueRecommendation) {
	t.sb.WriteString(string(recommendedValueType))
}

func (t *testTypeVisitor) String() string { return t.sb.String() }

func TestRecommendedTypeString(t *testing.T) {
	p := NewParser()
	tree, err := p.Parse("title==foo*")
	assert.NoError(t, err)
	v := &testTypeVisitor{}
	tree.Accept(v)
	assert.Equal(t, "string", v.String())
}

func TestRecommendedTypeStrings(t *testing.T) {
	p := NewParser()
	tree, err := p.Parse("title==foo,blah==blub")
	assert.NoError(t, err)
	v := &testTypeVisitor{}
	tree.Accept(v)
	assert.Equal(t, "stringstring", v.String())
}

func TestRecommendedTypeTuple(t *testing.T) {
	p := NewParser()
	tree, err := p.Parse("title=in=[a*b]")
	assert.NoError(t, err)
	v := &testTypeVisitor{}
	tree.Accept(v)
	assert.Equal(t, "tuple", v.String())
}

func TestRecommendedTypeNumber(t *testing.T) {
	p := NewParser()
	tree, err := p.Parse("title=lt=-100.0")
	assert.NoError(t, err)
	v := &testTypeVisitor{}
	tree.Accept(v)
	assert.Equal(t, "number", v.String())
}

func TestRecommendedTypeDateTime(t *testing.T) {
	p := NewParser()
	tree, err := p.Parse("title=gt=2003-12-13T18:30:02Z")
	assert.NoError(t, err)
	v := &testTypeVisitor{}
	tree.Accept(v)
	assert.Equal(t, "datetime", v.String())
}

func TestRecommendedTypeDuration(t *testing.T) {
	p := NewParser()
	tree, err := p.Parse("title=gt=-P5Y")
	assert.NoError(t, err)
	v := &testTypeVisitor{}
	tree.Accept(v)
	assert.Equal(t, "duration", v.String())
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
