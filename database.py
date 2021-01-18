#!/usr/bin/env python3

"""
database.py

A series of functions to leverage the (node, edge) schema of 
json-based nodes, and edges with optional json properties,
using an atomic transaction wrapper function.

"""

import sqlite3
import json
from itertools import tee
import os
import uuid
modulepath = os.path.dirname(__file__)

def uuidstr2uuidint(uuidhex):
    uid = uuid.UUID(uuidhex)
    return uid.int

def uuidint2uuidhex(uuidint):
    return uuid.UUID(int=uuidint).hex

def atomic(db_file, cursor_exec_fn):
    connection = sqlite3.connect(db_file)
    cursor = connection.cursor()
    cursor.execute("PRAGMA foreign_keys = TRUE;")
    results = cursor_exec_fn(cursor)
    connection.commit()
    connection.close()
    return results

def initialize(db_file, schema_file=None):
    if schema_file is None:
        schema_file = os.path.join(modulepath, 'schema.sql')

    def _init(cursor):
        with open(schema_file) as f:
            for statement in f.read().split(';'):
                cursor.execute(statement)
    return atomic(db_file, _init)

def _set_id(identifier, data):
    if identifier is not None:
        data["id"] = identifier
    return data

def _insert_node(cursor, identifier, data):
    cursor.execute("INSERT INTO nodes VALUES(json(?))", (json.dumps(_set_id(identifier, data)),))

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
            cursor.execute("UPDATE nodes SET body = json(?) WHERE id = ?", (json.dumps(_set_id(identifier, updated_data)), identifier,))
    return _upsert_node

def connect_nodes(source_id, target_id, properties={}):
    def _connect_nodes(cursor):
        cursor.execute("INSERT INTO edges VALUES(?, ?, json(?))", (source_id, target_id, json.dumps(properties),))
    return _connect_nodes

def remove_node(identifier):
    def _remove_node(cursor):
        cursor.execute("DELETE FROM edges WHERE source = ? OR target = ?", (identifier, identifier,))
        cursor.execute("DELETE FROM nodes WHERE id = ?", (identifier,))
    return _remove_node

def _parse_search_results(results, idx=0):
    presults = []
    for item in results:
        try:
            presults.append(json.loads(item[idx]))
        except:
            presults.append(item[idx])
    return presults

def find_node(identifier):
    def _find_node(cursor):
        results = cursor.execute("SELECT body FROM nodes WHERE json_extract(body, '$.id') = ?", (identifier,)).fetchall()
        # results = cursor.execute("SELECT body FROM nodes WHERE json_extract(body, '$.id') = {}".format(identifier)).fetchall()
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
        return _parse_search_results(cursor.execute("SELECT body FROM nodes WHERE {}".format(where_fn(data)), search_fn(data)).fetchall())
    return _find_nodes

def find_neighbors(identifier):
    def _find_neighbors(cursor):
        return cursor.execute("SELECT * FROM edges WHERE source = ? OR target = ?", (identifier, identifier,)).fetchall()
    return _find_neighbors

def find_outbound_neighbors(identifier):
    def _find_outbound_neighbors(cursor):
        return cursor.execute("SELECT * FROM edges WHERE source = ?", (identifier,)).fetchall()
    return _find_outbound_neighbors

def find_inbound_neighbors(identifier):
    def _find_inbound_neighbors(cursor):
        return cursor.execute("SELECT * FROM edges WHERE target = ?", (identifier,)).fetchall()
    return _find_inbound_neighbors

def _get_edge_sources(results):
    return _parse_search_results(results, 0)

def _get_edge_targets(results):
    return _parse_search_results(results, 1)

def _get_edge_properties(results):
    return _parse_search_results(results, 2)

def traverse (db_file, src, tgt=None, neighbors_fn=find_neighbors):
    def _depth_first_search(cursor):
        path = []
        queue = []
        if atomic(db_file, find_node(src)):
            queue.append(src)
        while queue:
            node = queue.pop()
            if node not in path:
                path.append(node)
                if node == tgt:
                    break
                neighbors = atomic(db_file, neighbors_fn(node))
                for identifier in set(_get_edge_sources(neighbors)).union(_get_edge_targets(neighbors)):
                    neighbor = atomic(db_file, find_node(identifier))
                    if neighbor:
                        queue.append(identifier)
        return path
    return atomic(db_file, _depth_first_search)

def get_connections(source_id, target_id):
    def _get_connections(cursor):
        return cursor.execute("SELECT * FROM edges WHERE source = ? AND target = ?", (source_id, target_id,)).fetchall()
    return _get_connections

def get_source_connections(source_id):
    def _get_connections(cursor):
        return cursor.execute("SELECT * FROM edges WHERE source = ? ", (source_id, )).fetchall()
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
    # sanitize keys (remove "!")
    sbody = {}
    for k, v in body.items():
        sbody[k.replace("!", "")] = v

    values = _fstring_from_keys([k for k in sbody.keys() if k not in exclude_keys], hide_key_name, kv_separator).format(**sbody)
    return f"[label=\"{values}\"]"

def _as_dot_node(body, exclude_keys=[], hide_key_name=False, kv_separator=' '):
    name = body['id']
    exclude_keys.append('id')
    label = _as_dot_label(body, exclude_keys, hide_key_name, kv_separator)
    return f'"{name}" {label};\n'

def _as_dot_edge(src, tgt, body, exclude_keys=[], hide_key_name=False, kv_separator=' '):
    label = _as_dot_label(body, exclude_keys, hide_key_name, kv_separator)
    return f'"{src}" -> "{tgt}" {label};\n'

def visualize(db_file, dot_file=None, path=[], \
        exclude_node_keys=[], hide_node_key=False, node_kv=' ', \
        exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):
    if dot_file is None:
        dot_file = "simple_graph.dot"
    def _visualize(cursor):
        with open(dot_file, 'w') as w:
            w.write("digraph {\n")
            for node in [atomic(db_file, find_node(i)) for i in path]:
                w.write(_as_dot_node(node, exclude_node_keys, hide_node_key, node_kv))
            for src, tgt in pairwise(path):
                for inbound in _get_edge_properties(atomic(db_file, get_connections(src, tgt))):
                    w.write(_as_dot_edge(src, tgt, inbound, exclude_edge_keys, hide_edge_key, edge_kv))
                for outbound in _get_edge_properties(atomic(db_file, get_connections(tgt, src))):
                    w.write(_as_dot_edge(tgt, src, outbound, exclude_edge_keys, hide_edge_key, edge_kv))
            w.write("}\n")
    return atomic(db_file, _visualize)

def get_dot(db_file, dot_file=None, path=[], \
        exclude_node_keys=[], hide_node_key=False, node_kv=' ', \
        exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):
    def _visualize(cursor):
        dots = []
        dots.append("digraph {\n")
        for node in [atomic(db_file, find_node(i)) for i in path]:
            dots.append(_as_dot_node(node, exclude_node_keys, hide_node_key, node_kv))
            src = node["uuid"]
            for src, tgt, inbound in atomic(db_file, get_source_connections(src)):
                dots.append(_as_dot_edge(src, tgt, {}, exclude_edge_keys, hide_edge_key, edge_kv))
        dots.append("}\n")
        if dot_file is not None:
            with open(dot_file, 'w') as fh:
                fh.writelines(dots)
        return "".join(dots)
    return atomic(db_file, _visualize)
