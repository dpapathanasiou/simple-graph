# -*-coding: utf-8-*-
import sys
import os
import uuid
import copy

modulepath = os.path.dirname(__file__)
sys.path.insert(0, os.path.join(modulepath, "..", ))

from simple_graph_db.db import Database

class Node():
    def __init__(self, name="N.N."):
        self.name = name
        self.uuid = uuid.uuid4().hex
        self.parent = None

    def copy(self, name=None):
        new_node = copy.deepcopy(self)
        new_node.uuid = uuid.uuid4().hex
        new_node.parent = self.uuid
        if name is not None:
            new_node.name = name
        return new_node

    @property
    def data(self):
        return {"name":self.name,
                "uuid":self.uuid,
                "!parent":self.parent}

    def __str__(self):
        return "{0.name}:{0.uuid}".format(self)

    __repr__ = __str__

def test_database():
    db_file = "/tmp/simple_graph_db.sqlite"
    if os.path.exists(db_file):
        os.remove(db_file)

    db = Database(db_file=db_file, schema_file='schema_childs.sql')
    print(db)

    uid = "e59e12cca729483f969ad1feb1b1d17e"
    db.add_node(data={"a":1}, identifier=uid)
    ruid = db.find_node(uid)
    print(ruid)
    assert ruid.get("id") == uid

    uid2 = "afeeb876a7524c9f8f86af73e95f3785"
    db.add_node(data={"b":1}, identifier=uid2)
    ruid2 = db.find_node(uid2)
    print(ruid2)
    assert ruid2.get("b") == 1

    db.connect_nodes(uid, uid2, {"con":1})

    db.find_nodes({"b":1})

    find_neighbors = db.find_neighbors(uid)
    print(find_neighbors)

def test_uuid():
    nodes = []

    n0 = Node("n0")
    n01 = n0.copy("n01")
    n02 = n0.copy("n02")
    n011 = n01.copy("n011")
    n012 = n01.copy("n012")
    print(n0)
    print(n01)
    nodes.extend([n0, n01, n02, n011, n012])
    print(nodes)


    db_file = "/tmp/test_uuid_child.sqlite"
    if os.path.exists(db_file):
        os.remove(db_file)

    db = Database(db_file=db_file)
    print(db)

    for node in nodes:
        db.add_node(node.uuid, node.data)

    rn0 = db.find_node(n0.uuid)
    print(rn0)
    assert n0.uuid == rn0.get("uuid")

    rn011 = db.find_node(n011.uuid)
    print(rn011)
    assert n011.uuid == rn011.get("uuid")
    assert n011.parent == rn011.get("!parent")

    ids = [n.uuid for n in nodes]
    print(ids)

    db.connect_nodes(n0.uuid, n01.uuid, {'con': 1})
    db.connect_nodes(n01.uuid, n011.uuid, {'con': 2})
    db.connect_nodes(n01.uuid, n012.uuid, {'con': 3}, edge_type="link")

    # dotstr = db.get_dot(db_file, path=ids)
    # print("!", dotstr)

    find_neighbors = db.find_neighbors(n01.uuid)
    print(find_neighbors)
