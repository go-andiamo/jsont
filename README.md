# JSONT

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
