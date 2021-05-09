#!/usr/bin/env python3

"""
database.py

A series of functions to leverage the (node, edge) schema of 
json-based nodes, and edges with optional json properties,
using an atomic transaction wrapper function.

"""

import sqlite3
import json
import pathlib
from functools import lru_cache
from itertools import tee
from graphviz import Digraph

@lru_cache(maxsize=None)
def read_sql(sql_file):
    with open(pathlib.Path.cwd() / "sql" / sql_file) as f:
        return f.read()

def atomic(db_file, cursor_exec_fn):
    connection = sqlite3.connect(db_file)
    cursor = connection.cursor()
    cursor.execute("PRAGMA foreign_keys = TRUE;")
    results = cursor_exec_fn(cursor)
    connection.commit()
    connection.close()
    return results

def initialize(db_file, schema_file='schema.sql'):
    def _init(cursor):
        cursor.executescript(read_sql(schema_file))
    return atomic(db_file, _init)

def _set_id(identifier, data):
    if identifier is not None:
        data["id"] = identifier
    return data

def _insert_node(cursor, identifier, data):
    cursor.execute(read_sql('insert-node.sql'), (json.dumps(_set_id(identifier, data)),))

def add_node(data, identifier=None):
    def _add_node(cursor):
        _insert_node(cursor, identifier, data)
    return _add_node

def upsert_node(identifier, data):
    def _upsert_node(cursor):
        current_data = find_node(identifier)(cursor)
        if not current_data:
            # no prior record exists, so regular insert
            _insert_node(cursor, identifier, data)
        else:
            # merge the current and new data and update
            updated_data = {**current_data, **data}
            cursor.execute(read_sql('update-node.sql'), (json.dumps(_set_id(identifier, updated_data)), identifier,))
    return _upsert_node

def connect_nodes(source_id, target_id, properties={}):
    def _connect_nodes(cursor):
        cursor.execute(read_sql('insert-edge.sql'), (source_id, target_id, json.dumps(properties),))
    return _connect_nodes

def remove_node(identifier):
    def _remove_node(cursor):
        cursor.execute(read_sql('delete-edge.sql'), (identifier, identifier,))
        cursor.execute(read_sql('delete-node.sql'), (identifier,))
    return _remove_node

def _parse_search_results(results, idx=0):
    return [json.loads(item[idx]) for item in results]

def _get_edge_properties(results):
    return _parse_search_results(results, 2)

def find_node(identifier):
    def _find_node(cursor):
        results = cursor.execute(read_sql('search-node-by-id.sql'), (identifier,)).fetchall()
        if len(results) == 1:
            return _parse_search_results(results).pop()
        return {}
    return _find_node

def _search_where(properties, predicate='='):
    return " AND ".join(["json_extract(body, '$.{}') {} ?".format(k, predicate) for k in properties.keys()])

def _search_like(properties):
    return _search_where(properties, 'LIKE')

def _search_equals(properties):
    return tuple([str(v) for v in properties.values()])

def _search_starts_with(properties):
    return tuple([str(v)+'%' for v in properties.values()])

def _search_contains(properties):
    return tuple(['%'+str(v)+'%' for v in properties.values()])

def find_nodes(data, where_fn=_search_where, search_fn=_search_equals):
    def _find_nodes(cursor):
        return _parse_search_results(cursor.execute(read_sql('search-node.sql') + "{}".format(where_fn(data)), search_fn(data)).fetchall())
    return _find_nodes

def find_neighbors():
    return read_sql('traverse.sql')

def find_outbound_neighbors():
    return read_sql('traverse-outbound.sql')

def find_inbound_neighbors():
    return read_sql('traverse-inbound.sql')

def traverse (db_file, src, tgt=None, neighbors_fn=find_neighbors):
    def _traverse(cursor):
        path = []
        target = json.dumps(tgt)
        for row in cursor.execute(neighbors_fn(), (json.dumps(src,))):
            if row:
                identifier = row[0]
                if identifier not in path:
                    path.append(identifier)
                if identifier == target:
                    break
        return path
    return atomic(db_file, _traverse)

def get_connections(source_id, target_id):
    def _get_connections(cursor):
        return cursor.execute(read_sql('search-edges.sql'), (source_id, target_id,)).fetchall()
    return _get_connections

def pairwise(iterable):
    a, b = tee(iterable)
    next(b, None)
    return zip(a, b)

def _fstring_from_keys(keys, hide_key, kv_separator):
    if hide_key:
        return '\\n'.join(['{'+k+'}' for k in keys])
    return '\\n'.join([k+kv_separator+'{'+k+'}' for k in keys])

def _as_dot_label(body, exclude_keys, hide_key_name, kv_separator):
    return _fstring_from_keys([k for k in body.keys() if k not in exclude_keys], hide_key_name, kv_separator).format(**body)

def _as_dot_node(body, exclude_keys=[], hide_key_name=False, kv_separator=' '):
    name = body['id']
    exclude_keys.append('id')
    label = _as_dot_label(body, exclude_keys, hide_key_name, kv_separator)
    return str(name), label

def visualize(db_file, dot_file, path=[], format='png', \
        exclude_node_keys=[], hide_node_key=False, node_kv=' ', \
        exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):
    def _visualize(cursor):
        dot = Digraph()
        for node in [atomic(db_file, find_node(i)) for i in path]:
            name, label = _as_dot_node(node, exclude_node_keys, hide_node_key, node_kv)
            dot.node(name, label=label)
        for src, tgt in pairwise(path):
            for inbound in _get_edge_properties(atomic(db_file, get_connections(src, tgt))):
                dot.edge(str(src), str(tgt), label=_as_dot_label(inbound, exclude_edge_keys, hide_edge_key, edge_kv))
            for outbound in _get_edge_properties(atomic(db_file, get_connections(tgt, src))):
                dot.edge(str(tgt), str(src), label=_as_dot_label(outbound, exclude_edge_keys, hide_edge_key, edge_kv))
        dot.render(dot_file, format=format)
    return atomic(db_file, _visualize)
