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
from graphviz import Digraph


@lru_cache(maxsize=None)
def read_sql(sql_file):
    with open(pathlib.Path.cwd() / ".." / "sql" / sql_file) as f:
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
    cursor.execute(read_sql('insert-node.sql'),
                   (json.dumps(_set_id(identifier, data)),))


def add_node(data, identifier=None):
    def _add_node(cursor):
        _insert_node(cursor, identifier, data)
    return _add_node


def add_nodes(nodes, ids):
    def _add_nodes(cursor):
        cursor.executemany(read_sql('insert-node.sql'), [(x,) for x in map(
            lambda node: json.dumps(_set_id(node[0], node[1])), zip(ids, nodes))])
    return _add_nodes


def _upsert_node(cursor, identifier, data):
    current_data = find_node(identifier)(cursor)
    if not current_data:
        # no prior record exists, so regular insert
        _insert_node(cursor, identifier, data)
    else:
        # merge the current and new data and update
        updated_data = {**current_data, **data}
        cursor.execute(read_sql(
            'update-node.sql'), (json.dumps(_set_id(identifier, updated_data)), identifier,))


def upsert_node(identifier, data):
    def _upsert(cursor):
        _upsert_node(cursor, identifier, data)
    return _upsert


def upsert_nodes(nodes, ids):
    def _upsert(cursor):
        for (id, node) in zip(ids, nodes):
            _upsert_node(cursor, id, node)
    return _upsert


def connect_nodes(source_id, target_id, properties={}):
    def _connect_nodes(cursor):
        cursor.execute(read_sql('insert-edge.sql'),
                       (source_id, target_id, json.dumps(properties),))
    return _connect_nodes


def connect_many_nodes(sources, targets, properties):
    def _connect_nodes(cursor):
        cursor.executemany(read_sql(
            'insert-edge.sql'), [(x[0], x[1], json.dumps(x[2]),) for x in zip(sources, targets, properties)])
    return _connect_nodes


def remove_node(identifier):
    def _remove_node(cursor):
        cursor.execute(read_sql('delete-edge.sql'), (identifier, identifier,))
        cursor.execute(read_sql('delete-node.sql'), (identifier,))
    return _remove_node


def remove_nodes(identifiers):
    def _remove_node(cursor):
        cursor.executemany(read_sql(
            'delete-edge.sql'), [(identifier, identifier,) for identifier in identifiers])
        cursor.executemany(read_sql('delete-node.sql'),
                           [(identifier,) for identifier in identifiers])
    return _remove_node


def _parse_search_results(results, idx=0):
    return [json.loads(item[idx]) for item in results]


def find_node(identifier):
    def _find_node(cursor):
        result = cursor.execute(
            read_sql('search-node-by-id.sql'), (identifier,)).fetchone()
        return {} if not result else json.loads(result[0])
    return _find_node


def _search_where(properties, predicate='='):
    return " AND ".join([f"json_extract(body, '$.{key}') {predicate} ?" for key in properties.keys()])


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
        return _parse_search_results(cursor.execute(read_sql('search-node.sql') + where_fn(data), search_fn(data)).fetchall())
    return _find_nodes


def find_neighbors(with_bodies=False):
    return read_sql("traverse-with-bodies.sql") if with_bodies else read_sql('traverse.sql')


def find_outbound_neighbors(with_bodies=False):
    return read_sql("traverse-with-bodies-outbound.sql") if with_bodies else read_sql('traverse-outbound.sql')


def find_inbound_neighbors(with_bodies=False):
    return read_sql("traverse-with-bodies-inbound.sql") if with_bodies else read_sql('traverse-inbound.sql')


def traverse(db_file, src, tgt=None, neighbors_fn=find_neighbors):
    def _traverse(cursor):
        path = []
        target = json.dumps(tgt)
        for row in cursor.execute(neighbors_fn(), {'source': src}):
            if row:
                identifier = row[0]
                if identifier not in path:
                    path.append(identifier)
                if identifier == target:
                    break
        return path
    return atomic(db_file, _traverse)


def traverse_with_bodies(db_file, src, tgt=None, neighbors_fn=find_neighbors):
    def _traverse(cursor):
        path = []
        target = json.dumps(tgt)
        header = None
        for row in cursor.execute(neighbors_fn(True), {'source': src}):
            if not header:
                header = row
                continue
            if row:
                identifier, obj, _ = row
                path.append(row)
                if identifier == target and obj == '()':
                    break
        return path
    return atomic(db_file, _traverse)


def connections_in():
    return read_sql('search-edges-inbound.sql')


def connections_out():
    return read_sql('search-edges-outbound.sql')


def get_connections_one_way(identifier, direction=connections_in):
    def _get_connections(cursor):
        return cursor.execute(direction(), (identifier,)).fetchall()
    return _get_connections


def get_connections(identifier):
    def _get_connections(cursor):
        return cursor.execute(read_sql('search-edges.sql'), (identifier, identifier,)).fetchall()
    return _get_connections


def _as_dot_label(body, exclude_keys, hide_key_name, kv_separator):
    keys = [k for k in body.keys() if k not in exclude_keys]
    fstring = '\\n'.join(['{'+k+'}' for k in keys]) if hide_key_name else '\\n'.join(
        [k+kv_separator+'{'+k+'}' for k in keys])
    return fstring.format(**body)


def _as_dot_node(body, exclude_keys=[], hide_key_name=False, kv_separator=' '):
    name = body['id']
    exclude_keys.append('id')
    label = _as_dot_label(body, exclude_keys, hide_key_name, kv_separator)
    return str(name), label


def visualize(db_file, dot_file, path=[], connections=get_connections, format='png',
              exclude_node_keys=[], hide_node_key=False, node_kv=' ',
              exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):

    def _unpack_edge(edge):
        return [json.loads(item) for item in edge]

    ids = []
    for i in path:
        ids.append(i)
        for edge in atomic(db_file, connections(i)):
            src, tgt, _ = _unpack_edge(edge)
            if src not in ids:
                ids.append(src)
            if tgt not in ids:
                ids.append(tgt)

    dot = Digraph()

    visited = []
    edges = []
    for i in ids:
        if i not in visited:
            node = atomic(db_file, find_node(i))
            name, label = _as_dot_node(
                node, exclude_node_keys, hide_node_key, node_kv)
            dot.node(name, label=label)
            for edge in atomic(db_file, connections(i)):
                if edge not in edges:
                    src, tgt, props = _unpack_edge(edge)
                    dot.edge(str(src), str(tgt), label=_as_dot_label(
                        props, exclude_edge_keys, hide_edge_key, edge_kv) if props else None)
                    edges.append(edge)
            visited.append(i)

    dot.render(dot_file, format=format)


def visualize_bodies(dot_file, path=[], format='png',
                     exclude_node_keys=[], hide_node_key=False, node_kv=' ',
                     exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):
    dot = Digraph()
    current_id = None
    edges = []
    for (identifier, obj, properties) in path:
        body = json.loads(properties)
        if obj == '()':
            name, label = _as_dot_node(
                body, exclude_node_keys, hide_node_key, node_kv)
            dot.node(name, label=label)
            current_id = body['id']
        else:
            edge = (str(current_id), str(
                identifier), body) if obj == '->' else (str(identifier), str(current_id), body)
            if edge not in edges:
                dot.edge(edge[0], edge[1], label=_as_dot_label(
                    body, exclude_edge_keys, hide_edge_key, edge_kv) if body else None)
                edges.append(edge)
    dot.render(dot_file, format=format)
