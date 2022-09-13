package fiqlparser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var values = []struct {
		fiql        string
		stringOuput string
		errorOutput error
	}{
		{fiql: "column=q=value", stringOuput: "column query value", errorOutput: nil},
		{fiql: "column==value", stringOuput: "column == value", errorOutput: nil},
		{fiql: "updated=gt=2003-12-13T00:00:00Z", stringOuput: "updated > 2003-12-13T00:00:00Z", errorOutput: nil},
		{fiql: "column==value,b==a", stringOuput: "column == value OR b == a", errorOutput: nil},
		{fiql: "column==va\\,lue,b==a", stringOuput: "column == va,lue OR b == a", errorOutput: nil},
		{fiql: "title==foo*;(updated=lt=-P1D,title==*bar)", stringOuput: "title == foo* AND (updated < -P1D OR title == *bar )", errorOutput: nil},
		{fiql: "(title==foo*);(fml==x,(xfs==a;f==fx))", stringOuput: "(title == foo* )AND (fml == x OR (xfs == a AND f == fx ))", errorOutput: nil},
		{fiql: "(title==foo*,test==a,fx==fa);(fml==x)", stringOuput: "(title == foo* OR test == a OR fx == fa )AND (fml == x )", errorOutput: nil},
		{fiql: "(title==foo*);(fml==x,(xfs==a;f==fx)", stringOuput: "", errorOutput: errors.New("ln:1:36 syntax error (unclosed brace))")},
		{fiql: "title=ffoo*", stringOuput: "", errorOutput: errors.New("ln:1:6 unexpected input (got `=f` but expected one of ==,!=,=gt=,=ge=,=lt=,=le=,=in=,=q=)")},
		{fiql: "title==fo,o*", stringOuput: "", errorOutput: errors.New("ln:1:12 syntax error (got `eof` but expected a value)")},
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
