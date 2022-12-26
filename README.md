# About

This is a simple [graph database](https://en.wikipedia.org/wiki/Graph_database) in [SQLite](https://www.sqlite.org/), inspired by "[SQLite as a document database](https://dgl.cx/2020/06/sqlite-json-support)".

# Structure

The [schema](sql/schema.sql) consists of just two structures:

* Nodes - these are any [json](https://www.json.org/) objects, with the only constraint being that they each contain a unique `id` value 
* Edges - these are pairs of node `id` values, specifying the direction, with an optional json object as connection properties

There are also traversal functions as native SQLite [Common Table Expressions](https://www.sqlite.org/lang_with.html) which produce lists of identifiers or return all objects along the path:

* Both directions
  * [identifiers](sql/traverse.sql)
  * [all objects](sql/traverse-with-bodies.sql)
* Inbound
  * [identifiers](sql/traverse-inbound.sql)
  * [all objects](sql/traverse-with-bodies-inbound.sql)
* Outbound
  * [identifiers](sql/traverse-outbound.sql)
  * [all objects](sql/traverse-with-bodies-outbound.sql)

# Applications

* [Social networks](https://en.wikipedia.org/wiki/Social_graph)
* [Interest maps/recommendation finders](https://en.wikipedia.org/wiki/Interest_graph)
* [To-do / task lists](https://en.wikipedia.org/wiki/Task_list)
* [Bug trackers](https://en.wikipedia.org/wiki/Open-source_software_development#Bug_trackers_and_task_lists)
* [Customer relationship management (CRM)](https://en.wikipedia.org/wiki/Customer_relationship_management)
* [Gantt chart](https://en.wikipedia.org/wiki/Gantt_chart)

# Usage

## RESTful API (paid)

The [Banrai Simple Doc Store](https://banrai.net/) service wraps this database core with an API service, creating a no-admin database for both documents and graphs.

## Importable library (free)

Choose an implementation:

* [Python](python) (now [available in PyPI](https://pypi.org/project/simple-graph-sqlite/))
* [Go](go)
* [Julia](https://github.com/JuliaComputing/SQLiteGraph.jl) (courtesy of [Josh Day](https://github.com/joshday))
* [R](https://github.com/mikeasilva/simplegraphdb) (courtesy of [Michael Silva](https://github.com/mikeasilva))
* [Flutter and Dart](https://github.com/rodydavis/flutter_graph_database) (courtesy of [Rody Davis](https://github.com/rodydavis))

Want to contribute an implementation in your preferred programming language?

The [schema and prepared sql statements](sql) can be used by programs in *any* programming language with [SQLite bindings](https://en.wikipedia.org/wiki/SQLite#Programming_language_support). 

[Pull requests](https://help.github.com/articles/about-pull-requests/) are welcome!
