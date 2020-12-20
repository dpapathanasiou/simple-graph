# About

This is a simple [graph database](https://en.wikipedia.org/wiki/Graph_database) in [SQLite](https://www.sqlite.org/), inspired by "[SQLite as a document database](https://dgl.cx/2020/06/sqlite-json-support)".

# Structure

The [schema](schema.sql) consists of just two structures:

* Nodes - these are any [json](https://www.json.org/) objects, with the only constraint being that they each contain a unique `id` value 
* Edges - these are pairs of node `id` values, specifying the direction, with an optional json object as connection properties

# Applications

* [Social networks](https://en.wikipedia.org/wiki/Social_graph)
* [Interest maps/recommendation finders](https://en.wikipedia.org/wiki/Interest_graph)
* [To-do / task lists](https://en.wikipedia.org/wiki/Task_list)
* [Bug trackers](https://en.wikipedia.org/wiki/Open-source_software_development#Bug_trackers_and_task_lists)
* [Customer relationship management (CRM)](https://en.wikipedia.org/wiki/Customer_relationship_management)
* [Gantt chart](https://en.wikipedia.org/wiki/Gantt_chart)

# Usage

## Installation

* [SQLite](https://www.sqlite.org/)
* [Python](https://www.python.org/)

## Basic Functions

The python [database script](database.py) provides convenience functions for [atomic transactions](https://en.wikipedia.org/wiki/Atomicity_(database_systems)) to add, delete, connect, and search for nodes.

### Example

Dropping into a python shell, we can create and connect people from the early days of [Apple Computer](https://en.wikipedia.org/wiki/Apple_Inc.). The resulting database will be saved to a SQLite file named `apple.sqlite`:

```
>>> apple = "apple.sqlite"
>>> import database as db
>>> db.initialize(apple)
>>> db.atomic(apple, db.add_node({'name': 'Apple Computer Company', 'type':['company', 'start-up'], 'founded': 'April 1, 1976'}, 1))
>>> db.atomic(apple, db.add_node({'name': 'Steve Wozniak', 'type':['person','engineer','founder']}, 2))
>>> db.atomic(apple, db.add_node({'name': 'Steve Jobs', 'type':['person','designer','founder']}, 3))
>>> db.atomic(apple, db.add_node({'name': 'Ronald Wayne', 'type':['person','administrator','founder']}, 4))
>>> db.atomic(apple, db.add_node({'name': 'Mike Markkula', 'type':['person','investor']}, 5))
>>> db.atomic(apple, db.connect_nodes(2, 1, {'action': 'founded'}))
>>> db.atomic(apple, db.connect_nodes(3, 1, {'action': 'founded'}))
>>> db.atomic(apple, db.connect_nodes(4, 1, {'action': 'founded'}))
>>> db.atomic(apple, db.connect_nodes(5, 1, {'action': 'invested', 'equity': 80000, 'debt': 170000}))
>>> db.atomic(apple, db.connect_nodes(1, 4, {'action': 'divested', 'amount': 800, 'date': 'April 12, 1976'}))
>>> db.atomic(apple, db.connect_nodes(2, 3))
```
