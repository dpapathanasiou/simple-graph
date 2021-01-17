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

* [SQLite](https://www.sqlite.org/), version 3.31.0 or higher; get the latest source or precompiled binaries from the [SQLite Download Page](https://www.sqlite.org/download.html) 
* [Python](https://www.python.org/)
* [Graphviz](https://graphviz.org/) optional, for visualization

## Basic Functions

The python [database script](database.py) provides convenience functions for [atomic transactions](https://en.wikipedia.org/wiki/Atomicity_(database_systems)) to add, delete, connect, and search for nodes.

Any single node or path of nodes can also be depicted graphically by using the `visualize` function within the database script to generate [dot](https://graphviz.org/doc/info/lang.html) files, which in turn can be converted to images with Graphviz. 

## Testing

There will be more robust and dedicated unit tests with [pytest](https://docs.pytest.org/en/latest/) soon, but in the meantime, running the example locally will do in a pinch.

This bit of shell magic will pull out the commands from this document:

```sh
grep ">>> " README.md | grep -v "grep" | sed -e 's/>>> //'
```

Use a final `| clip` (Windows), `| pbcopy` (macOS), or `| xclip -selection clipboard` (most linuxes) to copy all the commands into your clipboard.

If you have the correct version of SQLite installed, everything should just work without errors.

### Example

Dropping into a python shell, we can create, [upsert](https://en.wiktionary.org/wiki/upsert), and connect people from the early days of [Apple Computer](https://en.wikipedia.org/wiki/Apple_Inc.). The resulting database will be saved to a SQLite file named `apple.sqlite`:

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
>>> db.atomic(apple, db.upsert_node(2, {'nickname': 'Woz'}))
```

The nodes can be searched by their ids or any other combination of attributes (either as strict equality, or using `_search_like` in combination with `_search_starts_with` or `_search_contains`):

```
>>> db.atomic(apple, db.find_node(1))
{'name': 'Apple Computer Company', 'type': ['company', 'start-up'], 'founded': 'April 1, 1976', 'id': 1}
>>> db.atomic(apple, db.find_nodes({'name': 'Steve'}, db._search_like, db._search_starts_with))
[{'name': 'Steve Wozniak', 'type': ['person', 'engineer', 'founder'], 'id': 2, 'nickname': 'Woz'}, {'name': 'Steve Jobs', 'type': ['person', 'designer', 'founder'], 'id': 3}]
```

Paths through the graph can be discovered with a starting node id, and an optional ending id; the default neighbor expansion is nodes connected nodes in either direction, but that can changed by specifying either `find_outbound_neighbors` or `find_inbound_neighbors` instead:

```
>>> db.traverse(apple, 2, 3)
[2, 3]
>>> db.traverse(apple, 4, 5)
[4, 1, 5]
>>> db.traverse(apple, 5, neighbors_fn=db.find_inbound_neighbors)
[5]
>>> db.traverse(apple, 5, neighbors_fn=db.find_outbound_neighbors)
[5, 1, 4]
>>> db.traverse(apple, 5, neighbors_fn=db.find_neighbors)
[5, 1, 4, 3, 2]
```

Any path or list of nodes can rendered graphically by using the `visualize` function. This command produces [dot](https://graphviz.org/doc/info/lang.html) files, which in turn can be converted to images with Graphviz:

```
>>> db.visualize(apple, 'apple.dot', [4, 1, 5])
```

The resuling file can produce a [png](https://en.wikipedia.org/wiki/Portable_Network_Graphics) image, using this command line instruction:

```sh
dot -Tpng apple.dot -o apple.png
```

The default options include every key/value pair (excluding the id) in the node and edge objects:

![Basic visualization](.examples/apple-raw.png)

There are display options to help refine what is produced:

```
>>> db.visualize(apple, 'apple.dot', [4, 1, 5], exclude_node_keys=['type'], hide_edge_key=True)
```

![More refined visualization](.examples/apple.png)

The resulting dot file can be edited further as needed; the [dot guide](https://graphviz.org/pdf/dotguide.pdf) has more options and examples.
