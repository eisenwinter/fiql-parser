<a name="readme-top"></a>

# fiql-parser

[![Go Report Card](https://goreportcard.com/badge/github.com/eisenwinter/fiql-parser)](https://goreportcard.com/report/github.com/eisenwinter/fiql-parser) [![Go](https://github.com/eisenwinter/fiql-parser/actions/workflows/go.yml/badge.svg)](https://github.com/eisenwinter/fiql-parser/actions/workflows/go.yml) [![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/eisenwinter/fiql-parser) [![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)

A [FIQL (Feed Item Query Language)](https://datatracker.ietf.org/doc/html/draft-nottingham-atompub-fiql-00) query parser written in golang.

## Getting Started

Install the package by using `go get`

```
go get github.com/eisenwinter/fiql-parser
```

Use it to parse your FIQL query as shown in this example 

```golang
package yourpackage

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

func Example() {
	p := fq.NewParser()
	tree, err := p.Parse("(title==foo*);(fml==x,(xfs==a;f==fx))")
	if err != nil {
		return
	}
	v := &SimpleTestVisitor{}
	tree.Accept(v)
	fmt.Print(v.String())
}

```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Why

This is an upstream dependency for [`fiql-sql-adapter`](https://github.com/eisenwinter/fiql-sql-adapter). I thought this might be useful in a standalone library at some point in time.


<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Issues, Questions, Recommendations

Just open an issue and I will get back to you.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Contributing

Although this is a simple upstream dependency for another library, I accept contributions.
The only thing to consider is not to break backward compatibility with the API.

If you have a suggestion that would make this better, please fork the repo and create a pull request. 

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## License

Distributed under the BSD-2-Clause license. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>