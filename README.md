# JSONT
[![GoDoc](https://godoc.org/github.com/go-andiamo/jsont?status.svg)](https://pkg.go.dev/github.com/go-andiamo/jsont)
[![Latest Version](https://img.shields.io/github/v/tag/go-andiamo/jsont.svg?sort=semver&style=flat&label=version&color=blue)](https://github.com/go-andiamo/jsont/releases)
[![codecov](https://codecov.io/gh/go-andiamo/jsont/branch/master/graph/badge.svg)](https://codecov.io/gh/go-andiamo/jsont)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-andiamo/jsont)](https://goreportcard.com/report/github.com/go-andiamo/jsont)
[![Maintainability](https://api.codeclimate.com/v1/badges/1d64bc6c8474c2074f2b/maintainability)](https://codeclimate.com/github/go-andiamo/jsont/maintainability)

A really simple JSON templating utility...

```go
package main

import "github.com/go-andiamo/jsont"

var myTemplate = jsont.MustCompileTemplate(`{
    "foo": ?,
    "bar": ?
}`)

func main() {
    str, _ := myTemplate.String("foo value", 1)
    println(str)
}
```
produces...
```json
{
    "foo": "foo value",
    "bar": 1
}
```

Or using named arg markers...
```go
package main

import "github.com/go-andiamo/jsont"

var myTemplate = jsont.MustCompileNamedTemplate(`{
    "foo": ?foo,
    "bar": ?bar
}`)

func main() {
    str, _ := myTemplate.String(map[string]interface{}{"foo": "foo value", "bar": 1})
    println(str)
}
```
