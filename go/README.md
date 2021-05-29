# About

This is the [Go](https://golang.org/) implementation.

## TODO

* Visualization in [Graphviz](https://graphviz.org/) using [go-graphviz](https://github.com/goccy/go-graphviz) (or equivalent)
* Marshalling/unmarshalling in [json](https://golang.org/pkg/encoding/json/) or [gabs](https://github.com/Jeffail/gabs) injecting "id" fields, merging bodies in update and [upsert](https://en.wiktionary.org/wiki/upsert)

# Usage

## Installation

* [SQLite](https://www.sqlite.org/), version 3.31.0 or higher; get the latest source or precompiled binaries from the [SQLite Download Page](https://www.sqlite.org/download.html) 
* [Go](https://golang.org/doc/install)
* [go-sqlite3](https://github.com/mattn/go-sqlite3)
  ```sh
  go get github.com/mattn/go-sqlite3
  ```
* Execute the [constants generation script](generate-constants.sh) to build the `constants.go` file using the statements in the [sql](../sql) folder
  ```sh
  ./generate-constants.sh
  ```

## Basic Functions

The [database package](simplegraph/database.go) provides convenience functions for [atomic transactions](https://en.wikipedia.org/wiki/Atomicity_(database_systems)) to add, delete, connect, and search for nodes.

## Testing

There are [unit tests](simplegraph/database_test.go) in the `simplegraph` package covering each of the basic functions.

Make sure to use the `json1` tags when running them:

```sh
cd simplegraph
go test -tags json1
```

If you have the correct version of SQLite installed, the tests should all pass:

```sh
PASS
ok  	github.com/dpapathanasiou/simple-graph/go/simplegraph	0.067s
```
