#!/usr/bin/env python3

"""
database.py

A series of functions to leverage the (node, edge) schema of 
json-based nodes, and edges with optional json properties,
using an atomic transaction wrapper function.

"""

import sqlite3
import json

def atomic(db_file, cursor_exec_fn):
    connection = sqlite3.connect(db_file)
    cursor = connection.cursor()
    results = cursor_exec_fn(cursor)
    connection.commit()
    connection.close()
    return results

def initialize(db_file, schema_file='schema.sql'):
    def _init(cursor):
        with open(schema_file) as f:
            for statement in f.read().split(';'):
                cursor.execute(statement)
    return atomic(db_file, _init)

def _set_id(identifier, data):
    if identifier is not None:
        data["id"] = identifier
    return data

def add_node(data, identifier=None):
    def _add_node(cursor):
        cursor.execute("INSERT INTO nodes VALUES(json(?))", (json.dumps(_set_id(identifier, data)),))
    return _add_node

def connect_nodes(source_id, target_id, properties={}):
    def _connect_nodes(cursor):
        cursor.execute("INSERT INTO edges VALUES(?, ?, json(?))", (source_id, target_id, json.dumps(properties),))
    return _connect_nodes

def remove_node(identifier):
    def _remove_node(cursor):
        cursor.execute("DELETE FROM edges WHERE source = ? OR target = ?", (identifier, identifier,))
        cursor.execute("DELETE FROM nodes WHERE id = ?", (identifier,))
    return _remove_node

def find_node(identifier):
    def _find_node(cursor):
        return cursor.execute("SELECT body FROM nodes WHERE json_extract(body, '$.id') = ?", (identifier,)).fetchall()
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
        return cursor.execute("SELECT body FROM nodes WHERE {}".format(where_fn(data)), search_fn(data)).fetchall()
    return _find_nodes
