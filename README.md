# About

This is a simple [graph database](https://en.wikipedia.org/wiki/Graph_database) in [SQLite](https://www.sqlite.org/), inspired by "[SQLite as a document database](https://dgl.cx/2020/06/sqlite-json-support)".

# Structure

The [schema](sql/schema.sql) consists of just two structures:

* Nodes - these are any [json](https://www.json.org/) objects, with the only constraint being that they each contain a unique `id` value 
* Edges - these are pairs of node `id` values, specifying the direction, with an optional json object as connection properties

There are also traversal functions as native SQLite [Common Table Expressions](https://www.sqlite.org/lang_with.html):

* [Both directions](sql/traverse.sql)
* [Inbound](sql/traverse-inbound.sql)
* [Outbound](sql/traverse-outbound.sql)

# Applications

* [Social networks](https://en.wikipedia.org/wiki/Social_graph)
* [Interest maps/recommendation finders](https://en.wikipedia.org/wiki/Interest_graph)
* [To-do / task lists](https://en.wikipedia.org/wiki/Task_list)
* [Bug trackers](https://en.wikipedia.org/wiki/Open-source_software_development#Bug_trackers_and_task_lists)
* [Customer relationship management (CRM)](https://en.wikipedia.org/wiki/Customer_relationship_management)
* [Gantt chart](https://en.wikipedia.org/wiki/Gantt_chart)

# Usage

Choose an implementation:

* [Python](python) (now [available in PyPI](https://pypi.org/project/simple-graph-sqlite/))
* [Go](go)

Want to contribute a version in your preferred language?

The [schema and prepared sql statements](sql) can be used by programs in *any* programming language with [SQLite bindings](https://en.wikipedia.org/wiki/SQLite#Programming_language_support). 

[Pull requests](https://help.github.com/articles/about-pull-requests/) are welcome!
