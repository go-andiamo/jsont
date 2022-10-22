// Package jsont - Go package for simple JSON templates
/*

A very simple way of building JSON strings (or []byte data) from a template

Create a template:
  jsonTemplate, _ := jsont.NewTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
(Note: escape arg position marker by using '??')

And then generate JSON from the template by supplying args:
  str, _ := jsonTemplate.String("aaa", true, 1.2)
  println(str)
would produce:
  {"foo":"aaa","bar":true,"baz":"?","qux":1.2}

Named arg templates can also be created and used:
  jsonTemplate, _ := jsont.NewNamedTemplate(`{"foo":?foo,"bar":?bar,"baz":"??","qux":?qux}`)
And then generate JSON from the template by supplying named args:
  str, _ := jsonTemplate.String(map[string]interface{}{"foo":"aaa", "bar":true, "qux":1.2})
  println(str)
would produce:
  {"foo":"aaa","bar":true,"baz":"?","qux":1.2}

*/
package jsont
