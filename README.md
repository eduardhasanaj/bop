# bop : A body parser

## Overview

A package that abstracts the process of filling database entities from Form values.
ParseModel return also a list of decoded columns which can be used in the database query operations.
It supports out of the box three encoding formats:
- application/json
- multipart/form-data
- x-www-form-urlencoded

## Install

```
go get github.com/eduardhasanaj/bop
```
