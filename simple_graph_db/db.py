import sqlite3
import json
from itertools import tee
import os
import uuid
import logging
from . import Database as Database0
modulepath = os.path.dirname(__file__)


class SimpleGraphException(Exception): pass

class Database(Database0):
    """A simple graph database

    .. code-block:: python

        db = Database(db_file="/tmp/simple_graph_db.sqlite")


    """

    SCHEMA_FILE = 'schema_childs.sql'

    def __init__(self, db_file, **kwargs):
        """A simple graph database

        :param root_path:
        """
        schema_file = 'schema_childs.sql'
        Database0.__init__(self, db_file, schema_file=schema_file)

    def connect_nodes(self, source_id, target_id, properties={}, edge_type=None):
        """connect_nodes

        :param source_id:
        :param target_id:
        :param properties:
        :return:
        """

        if edge_type is None:
            edge_type = "edge"
        return self.atomic(self.__connect_nodes(source_id, target_id, properties=properties, edge_type=edge_type))

    def __connect_nodes(self, source_id, target_id, properties={}, edge_type=None):
        """__connect_nodes

        :param source_id:
        :param target_id:
        :param properties:
        :return:
        """
        if edge_type is None:
            edge_type = "edge"

        def _connect_nodes(cursor):
            cursor.execute("INSERT INTO edges VALUES(?, ?, ?, json(?))", (source_id, target_id, edge_type, json.dumps(properties),))

        return _connect_nodes

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
                    for src, tgt, edge_type, inbound in self.atomic(self.__get_source_connections(src)):
                        dots.append(self._as_dot_edge(src, tgt, {}, exclude_edge_keys, hide_edge_key, edge_kv))
            dots.append("}\n")
            if dot_file is not None:
                with open(dot_file, 'w') as fh:
                    fh.writelines(dots)
            return "".join(dots)

        return self.atomic(_visualize)

    def __find_neighbors(self, identifier):
        """find_neighbors

        :param identifier:
        :return:
        """

        def _find_neighbors(cursor):
            return cursor.execute("SELECT * FROM edges WHERE source = ? OR target = ?",
                                  (identifier, identifier,)).fetchall()

        return _find_neighbors