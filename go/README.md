# About

This is the [Go](https://golang.org/) implementation.

# Usage

## Installation

* [SQLite](https://www.sqlite.org/), version 3.31.0 or higher; get the latest source or precompiled binaries from the [SQLite Download Page](https://www.sqlite.org/download.html) 
* [Go](https://golang.org/doc/install)
* [go-sqlite3](https://github.com/mattn/go-sqlite3)
  ```sh
  go get github.com/mattn/go-sqlite3
  ```
* [go-graphviz](https://github.com/goccy/go-graphviz) for visualizations
  ```sh
  go get github.com/goccy/go-graphviz
  ```
* Optionally, or if the statements in the [sql](../sql) folder change during development, run the [constants generation script](generate-constants.sh) to rebuild the `constants.go` file
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

### TODO

- [ ] Marshall/unmarshal using [json](https://golang.org/pkg/encoding/json/) (or [gabs](https://github.com/Jeffail/gabs), etc.) for injecting "id" fields, and merging bodies in update and [upsert](https://en.wiktionary.org/wiki/upsert) instead of the full replacement that happens now