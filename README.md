# bop : A body parser

## Overview

A package that abstracts the process of filling database entities from Form values or raw json.
ParseModel return also a list of decoded columns which can be used in the database query operations.
It supports out of the box three encoding formats:
- application/json
- multipart/form-data
- x-www-form-urlencoded

## Install

```
go get github.com/eduardhasanaj/bop
```

## Motivation
I was working on a e-commerce backend and it was required to have a flexible api: sending just partial 
properties which should be updated. For models with many propertis it was a tedious and repetitive task
to manually map  fields of the form to the appropriate struct field.
The code looks like:
```
columns := make([]string, 0)

custom := &models.Customer{}
field := r.PostForm("first_name")
if len(field) != 0 {
    custom.FirstName = field
    columns = append(columns, field)
}

field = r.PostForm("last_name")
if len(field) != 0 {
    custom.LastName = field
    columns = append(columns, field)
}

field = r.PostForm("username")
if len(field) != 0 {
    custom.Username = field
    columns = append(columns, field)
}

field = r.PostForm("password")
if len(field) != 0 {
    custom.password = field
    columns = append(columns, field)
}

field = r.PostForm("active")
if len(field) != 0 {
    custom.Active = field
    columns = append(columns, field)
}
```

With bop the above example is simplified to:
```
var custom models.Customer
p := bop.New(w, r)
columns, err := p.ParseModel(&custom)
```

## Benchmarks
```
BenchmarkParseModelPostForm-8   	 1374523	       864 ns/op	     160 B/op	       7 allocs/op
BenchmarkParseJsonModel-8   	 1671302	       707 ns/op	     768 B/op	       6 allocs/op
```