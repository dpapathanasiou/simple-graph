# -*-coding: utf-8-*-

__version__ = "0.1.0"

import sqlite3
import json
from itertools import tee
import os
import uuid
import logging

modulepath = os.path.dirname(__file__)


class SimpleGraphException(Exception): pass


class Database():
    """A simple graph database

    .. code-block:: python

        db = Database(db_file="/tmp/simple_graph_db.sqlite")


    """

    def __init__(self, db_file, **kwargs):
        """A simple graph database

        :param root_path:
        """

        self.db_file = db_file
        self.initialize(kwargs.get("schema_file"))

    def __str__(self):
        return "(DB:'{0}')".format(self.db_file)

    __repr__ = __str__

    def atomic(self, cursor_exec_fn):
        """

        :param cursor_exec_fn:
        :return:
        """
        connection = sqlite3.connect(self.db_file)
        try:
            cursor = connection.cursor()
            cursor.execute("PRAGMA foreign_keys = TRUE;")
            results = cursor_exec_fn(cursor)
            connection.commit()
            connection.close()
            return results

        except Exception as exp:
            logging.error(exp)
            connection.rollback()
            raise
        finally:
            connection.close()

    def initialize(self, schema_file=None):
        """

        :param schema_file:
        :return:
        """
        if schema_file is None:
            schema_file = os.path.join(modulepath, 'schema.sql')

        def _init(cursor):
            with open(schema_file) as f:
                for statement in f.read().split(';'):
                    cursor.execute(statement)

        return self.atomic(_init)

    def _set_id(self, identifier, data):
        """

        :param data:
        :return:
        """
        if identifier is not None:
            data["id"] = identifier
        return data

    def _insert_node(self, cursor, identifier, data):
        """

        :param cursor:
        :param identifier:
        :param data:
        :return:
        """
        cursor.execute("INSERT INTO nodes VALUES(json(?))", (json.dumps(self._set_id(identifier, data)),))

    def add_node(self, identifier=None, data=None):
        """add_node

        :param data:
        :param identifier:
        :return:
        """
        if data is None:
            data = {}
        if identifier is None:
            identifier = uuid.uuid4().hex

        try:
            return self.atomic(self.__add_node(data=data, identifier=identifier))
        except sqlite3.IntegrityError as exp:
            raise SimpleGraphException("nodes.id '{}' already in use. Try db.upsert_node()!".format(identifier))

    def __add_node(self, data=None, identifier=None):
        """__add_node

        :param data:
        :param identifier:
        :return:
        """

        def _add_node(cursor):
            self._insert_node(cursor, identifier, data)
            return identifier

        return _add_node

    def upsert_node(self, identifier=None, data=None):
        """upsert_node

        :param identifier:
        :param data:
        :return:
        """
        if data is None:
            data = {}
        if identifier is None:
            identifier = uuid.uuid4().hex
        return self.atomic(self.__upsert_node(identifier, data))

    def __upsert_node(self, identifier, data):
        """upsert_node

        :param identifier:
        :param data:
        :return:
        """

        def _upsert_node(cursor):
            current_data = self.__find_node(identifier)(cursor)
            if not current_data:
                # no prior record exists, so regular insert
                self._insert_node(cursor, identifier, data)
            else:
                # merge the current and new data and update
                updated_data = {**current_data, **data}
                cursor.execute("UPDATE nodes SET body = json(?) WHERE id = ?",
                               (json.dumps(self._set_id(identifier, updated_data)), identifier,))
            return identifier

        return _upsert_node

    def connect_nodes(self, source_id, target_id, properties={}):
        """connect_nodes

        :param source_id:
        :param target_id:
        :param properties:
        :return:
        """
        return self.atomic(self.__connect_nodes(source_id, target_id, properties))

    def __connect_nodes(self, source_id, target_id, properties={}):
        """__connect_nodes

        :param source_id:
        :param target_id:
        :param properties:
        :return:
        """

        def _connect_nodes(cursor):
            cursor.execute("INSERT INTO edges VALUES(?, ?, json(?))", (source_id, target_id, json.dumps(properties),))

        return _connect_nodes

    def remove_node(self, identifier):
        """remove_node('4480c7fcbffe4f6aa8db0b8e68d1f36f')

        :param identifier:
        :return:
        """
        return self.atomic(self.__remove_node(identifier))

    def __remove_node(self, identifier):
        """__remove_node

        :param identifier:
        :return:
        """

        def _remove_node(cursor):
            cursor.execute("DELETE FROM edges WHERE source = ? OR target = ?", (identifier, identifier,))
            cursor.execute("DELETE FROM nodes WHERE id = ?", (identifier,))

        return _remove_node

    def _parse_search_results(self, results, idx=0):
        """_parse_search_results

        :param results:
        :param idx:
        :return:
        """
        presults = []
        for item in results:
            try:
                presults.append(json.loads(item[idx]))
            except:
                presults.append(item[idx])
        return presults

    def find_node(self, identifier):
        """find_node('b641b0cf77844a24920fff3e6c71c8cf')

        :param identifier:
        :return:
        """
        return self.atomic(self.__find_node(identifier))

    def __find_node(self, identifier):
        """__find_node

        :param identifier:
        :return:
        """

        def _find_node(cursor):
            results = cursor.execute("SELECT body FROM nodes WHERE json_extract(body, '$.id') = ?",
                                     (identifier,)).fetchall()
            # results = cursor.execute("SELECT body FROM nodes WHERE json_extract(body, '$.id') = {}".format(identifier)).fetchall()
            if len(results) == 1:
                return self._parse_search_results(results).pop()
            return {}

        return _find_node

    def _search_where(self, properties, predicate='='):
        """_search_where

        :param properties:
        :param predicate:
        :return:
        """
        return " AND ".join(["json_extract(body, '$.{}') {} ?".format(k, predicate) for k in properties.keys()])

    def _search_like(self, properties):
        """_search_like

        :param properties:
        :return:
        """
        return self._search_where(properties, 'LIKE')

    def _search_equals(self, properties):
        """_search_equals

        :param properties:
        :return:
        """
        return tuple([str(v) for v in properties.values()])

    def _search_starts_with(self, properties):
        """_search_starts_with

        :param properties:
        :return:
        """
        return tuple([str(v) + '%' for v in properties.values()])

    def _search_contains(self, properties):
        """_search_contains

        :param properties:
        :return:
        """
        return tuple(['%' + str(v) + '%' for v in properties.values()])

    def find_nodes(self, data, where_fn=_search_where, search_fn=_search_equals):
        """find_nodes({'name': ''})

        :param data:
        :param where_fn:
        :param search_fn:
        :return:
        """
        return self.atomic(self.__find_nodes(data, self._search_like, self._search_starts_with))

    def __find_nodes(self, data, where_fn=_search_where, search_fn=_search_equals):
        """__find_nodes

        :param data:
        :param where_fn:
        :param search_fn:
        :return:
        """

        def _find_nodes(cursor):
            return self._parse_search_results(
                cursor.execute("SELECT body FROM nodes WHERE {}".format(where_fn(data)), search_fn(data)).fetchall())

        return _find_nodes

    def get_all_nodes(self):
        """get all nodes

        :return: list of node data
        """
        def _find_nodes(cursor):
            return self._parse_search_results(
                cursor.execute("SELECT body FROM nodes").fetchall())

        return self.atomic(_find_nodes)

    def get_all_node_ids(self):
        """get all node id's

        :return: list of ids
        """
        def _find_nodes(cursor):
            return self._parse_search_results(
                cursor.execute("SELECT id FROM nodes").fetchall())

        return self.atomic(_find_nodes)

    def get_all_edges(self):
        """get all edges

        :return: list of edge data
        """
        def _find_neighbors(cursor):
            return cursor.execute("SELECT * FROM edges").fetchall()

        return self.atomic(_find_neighbors)

    def find_neighbors(self, identifier):
        """find_neighbors

        :param identifier:
        :return:
        """
        return self.atomic(self.__find_neighbors(identifier))

    def __find_neighbors(self, identifier):
        """find_neighbors

        :param identifier:
        :return:
        """

        def _find_neighbors(cursor):
            return cursor.execute("SELECT * FROM edges WHERE source = ? OR target = ?",
                                  (identifier, identifier,)).fetchall()

        return _find_neighbors

    def find_outbound_neighbors(self, identifier):
        """find_outbound_neighbors

        :param identifier:
        :return:
        """

        def _find_outbound_neighbors(cursor):
            return cursor.execute("SELECT * FROM edges WHERE source = ?", (identifier,)).fetchall()

        return _find_outbound_neighbors

    def find_inbound_neighbors(self, identifier):
        """find_inbound_neighbors

        :param identifier:
        :return:
        """

        def _find_inbound_neighbors(cursor):
            return cursor.execute("SELECT * FROM edges WHERE target = ?", (identifier,)).fetchall()

        return _find_inbound_neighbors

    def _get_edge_sources(self, results):
        """_get_edge_sources

        :param results:
        :return:
        """
        return self._parse_search_results(results, 0)

    def _get_edge_targets(self, results):
        """_get_edge_targets

        :param results:
        :return:
        """
        return self._parse_search_results(results, 1)

    def _get_edge_properties(self, results):
        """_get_edge_properties

        :param results:
        :return:
        """
        return self._parse_search_results(results, 2)

    def traverse(self, src, tgt=None, neighbors_fn=__find_neighbors):
        """traverse

        :param src:
        :param tgt:
        :param neighbors_fn:
        :return:
        """

        def _depth_first_search(cursor):
            path = []
            queue = []
            if self.atomic(self.__find_node(src)):
                queue.append(src)
            while queue:
                node = queue.pop()
                if node not in path:
                    path.append(node)
                    if node == tgt:
                        break
                    neighbors = self.atomic(neighbors_fn(node))
                    for identifier in set(self._get_edge_sources(neighbors)).union(self._get_edge_targets(neighbors)):
                        neighbor = self.atomic(self.__find_node(identifier))
                        if neighbor:
                            queue.append(identifier)
            return path

        return self.atomic(_depth_first_search)

    def __get_connections(self, source_id, target_id):
        """__get_connections

        :param source_id:
        :param target_id:
        :return:
        """

        def _get_connections(cursor):
            return cursor.execute("SELECT * FROM edges WHERE source = ? AND target = ?",
                                  (source_id, target_id,)).fetchall()

        return _get_connections

    def __get_source_connections(self, source_id):
        """__get_source_connections

        :param source_id:
        :return:
        """

        def _get_connections(cursor):
            return cursor.execute("SELECT * FROM edges WHERE source = ? ", (source_id,)).fetchall()

        return _get_connections

    def pairwise(self, iterable):
        """

        :return:
        """
        a, b = tee(iterable)
        next(b, None)
        return zip(a, b)

    def _fstring_from_keys(self, keys, hide_key, kv_separator):
        """

        :param hide_key:
        :param kv_separator:
        :return:
        """
        if hide_key:
            return '\\n'.join(['{' + k + '}' for k in keys])
        return '\\n'.join([k + kv_separator + '{' + k + '}' for k in keys])

    def _as_dot_label(self, body, exclude_keys, hide_key_name, kv_separator):
        """

        :param exclude_keys:
        :param hide_key_name:
        :param kv_separator:
        :return:
        """
        # sanitize keys (remove "!")
        sbody = {}
        for k, v in body.items():
            sbody[k.replace("!", "")] = v

        values = self._fstring_from_keys([k for k in sbody.keys() if k not in exclude_keys], hide_key_name,
                                         kv_separator).format(**sbody)
        return f"[label=\"{values}\"]"

    def _as_dot_node(self, body, exclude_keys=[], hide_key_name=False, kv_separator=' '):
        """_as_dot_node

        :param body:
        :param exclude_keys:
        :param hide_key_name:
        :param kv_separator:
        :return:
        """
        name = body.get('id')
        if name is None:
            return '"?"'
        exclude_keys.append('id')
        label = self._as_dot_label(body, exclude_keys, hide_key_name, kv_separator)
        return f'"{name}" {label};\n'

    def _as_dot_edge(self, src, tgt, body, exclude_keys=[], hide_key_name=False, kv_separator=' '):
        """_as_dot_edge

        :param tgt:
        :param body:
        :param exclude_keys:
        :param hide_key_name:
        :param kv_separator:
        :return:
        """
        label = self._as_dot_label(body, exclude_keys, hide_key_name, kv_separator)
        return f'"{src}" -> "{tgt}" {label};\n'

    def visualize(self, dot_file=None, path=[], \
                  exclude_node_keys=[], hide_node_key=False, node_kv=' ', \
                  exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):
        """

        :param dot_file:
        :param path:
        :param exclude_node_keys:
        :param hide_node_key:
        :param node_kv:
        :param exclude_edge_keys:
        :param hide_edge_key:
        :param edge_kv:
        :return:
        """
        if dot_file is None:
            dot_file = "simple_graph.dot"

        def _visualize(cursor):
            with open(dot_file, 'w') as w:
                w.write("digraph {\n")
                for node in [self.atomic(self.__find_node(i)) for i in path]:
                    w.write(self._as_dot_node(node, exclude_node_keys, hide_node_key, node_kv))
                for src, tgt in self.pairwise(path):
                    for inbound in self._get_edge_properties(self.atomic(self.__get_connections(src, tgt))):
                        w.write(self._as_dot_edge(src, tgt, inbound, exclude_edge_keys, hide_edge_key, edge_kv))
                    for outbound in self._get_edge_properties(self.atomic(self.__get_connections(tgt, src))):
                        w.write(self._as_dot_edge(tgt, src, outbound, exclude_edge_keys, hide_edge_key, edge_kv))
                w.write("}\n")
            return dot_file

        return self.atomic(_visualize)

    def get_dot(self, dot_file=None, path=[], \
                exclude_node_keys=[], hide_node_key=False, node_kv=' ', \
                exclude_edge_keys=[], hide_edge_key=False, edge_kv=' '):
        """create a dot file

        :param dot_file:
        :param path:
        :param exclude_node_keys:
        :param hide_node_key:
        :param node_kv:
        :param exclude_edge_keys:
        :param hide_edge_key:
        :param edge_kv:
        :return: dot_str
        """

        def _visualize(cursor):
            dots = []
            dots.append("digraph {\n")
            for node in [self.atomic(self.__find_node(i)) for i in path]:
                dots.append(self._as_dot_node(node, exclude_node_keys, hide_node_key, node_kv))
                src = node.get("id")
                if src is not None:
                    for src, tgt, inbound in self.atomic(self.__get_source_connections(src)):
                        dots.append(self._as_dot_edge(src, tgt, {}, exclude_edge_keys, hide_edge_key, edge_kv))
            dots.append("}\n")
            if dot_file is not None:
                with open(dot_file, 'w') as fh:
                    fh.writelines(dots)
            return "".join(dots)

        return self.atomic(_visualize)
