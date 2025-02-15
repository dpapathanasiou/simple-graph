# About

This is a simple [graph database](https://en.wikipedia.org/wiki/Graph_database) in [SQLite](https://www.sqlite.org/), inspired by "[SQLite as a document database](https://dgl.cx/2020/06/sqlite-json-support)".

# Structure

The [schema](sql/schema.sql) consists of just two structures:

* Nodes - these are any [json](https://www.json.org/) objects, with the only constraint being that they each contain a unique `id` value 
* Edges - these are pairs of node `id` values, specifying the direction, with an optional json object as connection properties

The create, read, update, and delete functions ([.sql files](sql)) are complete statements with [qmark](https://docs.python.org/3/library/sqlite3.html#sqlite3.paramstyle) bindings.

Search templates ([.template files](sql)) are in [Jinja2](https://pypi.org/project/Jinja2/) format, which can be converted to other template syntaxes relatively easily, with a bit of [regex magic](https://github.com/dpapathanasiou/simple-graph-go/blob/main/generate-constants.sh) (though it would be nice if they could be expressed in a more language-agnostic way).

There are also traversal function templates as native SQLite [Common Table Expressions](https://www.sqlite.org/lang_with.html) which produce lists of identifiers or return all objects along the path.

# Applications

* [Social networks](https://en.wikipedia.org/wiki/Social_graph)
* [Interest maps/recommendation finders](https://en.wikipedia.org/wiki/Interest_graph)
* [To-do / task lists](https://en.wikipedia.org/wiki/Task_list)
* [Bug trackers](https://en.wikipedia.org/wiki/Open-source_software_development#Bug_trackers_and_task_lists)
* [Customer relationship management (CRM)](https://en.wikipedia.org/wiki/Customer_relationship_management)
* [Gantt chart](https://en.wikipedia.org/wiki/Gantt_chart)

# Usage

## RESTful API (paid)

The [Banrai Simple Graph Database and Document Store](https://banrai.com/) wraps this core logic with an API service, creating a managed graph database and document store, with additional features not found in any of the public bindings.

## Importable library (free)

Choose an implementation:

* [Python](https://github.com/dpapathanasiou/simple-graph-pypi) (now [available in PyPI](https://pypi.org/project/simple-graph-sqlite/))
* [Go](https://github.com/dpapathanasiou/simple-graph-go)
* [Julia](https://github.com/JuliaComputing/SQLiteGraph.jl) (courtesy of [Josh Day](https://github.com/joshday))
* [R](https://github.com/mikeasilva/simplegraphdb) (courtesy of [Michael Silva](https://github.com/mikeasilva))
* [Flutter and Dart](https://github.com/rodydavis/flutter_graph_database) (courtesy of [Rody Davis](https://github.com/rodydavis))
* [Swift](https://swiftpackageindex.com/Jomy10/SimpleGraph) (Courtesy of [Jonas Everaert](https://github.com/jomy10)

Want to contribute an implementation in your preferred programming language?

The [schema and prepared sql statements](sql) can be used by programs in *any* programming language with [SQLite bindings](https://en.wikipedia.org/wiki/SQLite#Programming_language_support). 

[Pull requests](https://help.github.com/articles/about-pull-requests/) are welcome!
