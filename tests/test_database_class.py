# -*-coding: utf-8-*-
import sys
import os

modulepath = os.path.dirname(__file__)
sys.path.insert(0, os.path.join(modulepath, "..", ))

from simple_graph_db import Database



def test_database():
    db_file = "/tmp/simple_graph_db.sqlite"
    if os.path.exists(db_file):
        os.remove(db_file)

    db = Database(db_file="/tmp/simple_graph_db.sqlite")
    print(db)

    uid = "e59e12cca729483f969ad1feb1b1d17e"
    db.add_node(data={"a":1}, identifier=uid)
    db.find_node(uid)

    uid2 = "afeeb876a7524c9f8f86af73e95f3785"
    db.add_node(data={"b":1}, identifier=uid2)
    db.find_node(uid2)

    db.connect_nodes(uid, uid2, {"con":1})

    assert 1==2
