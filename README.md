# JSONT
[![GoDoc](https://godoc.org/github.com/go-andiamo/jsont?status.svg)](https://pkg.go.dev/github.com/go-andiamo/jsont)
[![Latest Version](https://img.shields.io/github/v/tag/go-andiamo/jsont.svg?sort=semver&style=flat&label=version&color=blue)](https://github.com/go-andiamo/jsont/releases)
[![codecov](https://codecov.io/gh/go-andiamo/jsont/branch/main/graph/badge.svg?token=V4XVYR0A8G)](https://codecov.io/gh/go-andiamo/jsont)

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
