#!/usr/bin/env python3

"""
visualizers.py

Functions to enable visualizations of graph data, starting with graphviz,
and extensible to other libraries.

"""

from graphviz import Digraph
from database import atomic, find_node, get_connections
import json


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


def graphviz_visualize(db_file, dot_file, path=[], connections=get_connections, format='png',
                       exclude_node_keys=[], hide_node_key=False, node_kv=' ',
                       exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):

    ids = []
    for i in path:
        ids.append(str(i))
        for edge in atomic(db_file, connections(i)):
            src, tgt, _ = edge
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
                    src, tgt, prps = edge
                    props = json.loads(prps)
                    dot.edge(str(src), str(tgt), label=_as_dot_label(
                        props, exclude_edge_keys, hide_edge_key, edge_kv) if props else None)
                    edges.append(edge)
            visited.append(i)

    dot.render(dot_file, format=format)


def graphviz_visualize_bodies(dot_file, path=[], format='png',
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
