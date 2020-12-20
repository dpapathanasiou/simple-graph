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

The nodes can be searched by their ids or any other combination of attributes (either as strict equality, or using `_search_like` in combination with `_search_starts_with` or `_search_contains`):

```
>>> db.atomic(apple, db.find_node(1))
{'name': 'Apple Computer Company', 'type': ['company', 'start-up'], 'founded': 'April 1, 1976', 'id': 1}
>>> db.atomic(apple, db.find_nodes({'name': 'Steve'}, db._search_like, db._search_starts_with))
[{'name': 'Steve Wozniak', 'type': ['person', 'engineer', 'founder'], 'id': 2}, {'name': 'Steve Jobs', 'type': ['person', 'designer', 'founder'], 'id': 3}]
```

Paths through the graph can be discovered with a starting node id, and an optional ending id; the default neighbor expansion is nodes connected nodes in either direction, but that can changed by specifying either `find_outbound_neighbors` or `find_inbound_neighbors` instead:

```
>>> db.traverse(apple, 2, 3)
[{'name': 'Steve Wozniak', 'type': ['person', 'engineer', 'founder'], 'id': 2}, {'name': 'Steve Jobs', 'type': ['person', 'designer', 'founder'], 'id': 3}]
>>> db.traverse(apple, 4, 5)
[{'name': 'Ronald Wayne', 'type': ['person', 'administrator', 'founder'], 'id': 4}, {'name': 'Apple Computer Company', 'type': ['company', 'start-up'], 'founded': 'April 1, 1976', 'id': 1}, {'name': 'Mike Markkula', 'type': ['person', 'investor'], 'id': 5}]
>>> db.atomic(apple, db.find_inbound_neighbors(1))
[('2', '1', '{"action":"founded"}'), ('3', '1', '{"action":"founded"}'), ('4', '1', '{"action":"founded"}'), ('5', '1', '{"action":"invested","equity":80000,"debt":170000}')]
>>> db.atomic(apple, db.find_outbound_neighbors(1))
[('1', '4', '{"action":"divested","amount":800,"date":"April 12, 1976"}')]
>>> db.atomic(apple, db.find_outbound_neighbors(2))
[('2', '1', '{"action":"founded"}'), ('2', '3', '{}')]
>>> db.atomic(apple, db.find_inbound_neighbors(2))
[]
>>> db.atomic(apple, db.find_inbound_neighbors(3))
[('2', '3', '{}')]
>>> db.traverse(apple, 5, neighbors_fn=db.find_inbound_neighbors)
[{'name': 'Mike Markkula', 'type': ['person', 'investor'], 'id': 5}]
>>> db.traverse(apple, 5, neighbors_fn=db.find_outbound_neighbors)
[{'name': 'Mike Markkula', 'type': ['person', 'investor'], 'id': 5}, {'name': 'Apple Computer Company', 'type': ['company', 'start-up'], 'founded': 'April 1, 1976', 'id': 1}, {'name': 'Ronald Wayne', 'type': ['person', 'administrator', 'founder'], 'id': 4}]
>>> db.traverse(apple, 5, neighbors_fn=db.find_neighbors)
[{'name': 'Mike Markkula', 'type': ['person', 'investor'], 'id': 5}, {'name': 'Apple Computer Company', 'type': ['company', 'start-up'], 'founded': 'April 1, 1976', 'id': 1}, {'name': 'Ronald Wayne', 'type': ['person', 'administrator', 'founder'], 'id': 4}, {'name': 'Steve Jobs', 'type': ['person', 'designer', 'founder'], 'id': 3}, {'name': 'Steve Wozniak', 'type': ['person', 'engineer', 'founder'], 'id': 2}]
```
