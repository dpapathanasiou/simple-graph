# -*-coding: utf-8-*-
import sys
import os

modulepath = os.path.dirname(__file__)
sys.path.insert(0, os.path.join(modulepath, "..", ))

import database as db

def test_apple():
    db_path = "/tmp/apple.sqlite"
    if os.path.exists(db_path):
        os.remove(db_path)
    db.initialize(db_path)
    db.atomic(db_path, db.add_node(
        {'name': 'apple Computer Company', 'type': ['company', 'start-up'], 'founded': 'April 1, 1976'}, 1))
    db.atomic(db_path, db.add_node({'name': 'Steve Wozniak', 'type': ['person', 'engineer', 'founder']}, 2))
    db.atomic(db_path, db.add_node({'name': 'Steve Jobs', 'type': ['person', 'designer', 'founder']}, 3))
    db.atomic(db_path, db.add_node({'name': 'Ronald Wayne', 'type': ['person', 'administrator', 'founder']}, 4))
    db.atomic(db_path, db.add_node({'name': 'Mike Markkula', 'type': ['person', 'investor']}, 5))
    db.atomic(db_path, db.connect_nodes(2, 1, {'action': 'founded'}))
    db.atomic(db_path, db.connect_nodes(3, 1, {'action': 'founded'}))
    db.atomic(db_path, db.connect_nodes(4, 1, {'action': 'founded'}))
    db.atomic(db_path, db.connect_nodes(5, 1, {'action': 'invested', 'equity': 80000, 'debt': 170000}))
    db.atomic(db_path, db.connect_nodes(1, 4, {'action': 'divested', 'amount': 800, 'date': 'April 12, 1976'}))
    db.atomic(db_path, db.connect_nodes(2, 3))
    db.atomic(db_path, db.upsert_node(2, {'nickname': 'Woz'}))

    r1 = db.atomic(db_path, db.find_node(1))
    print(r1)
    assert r1["name"] == 'apple Computer Company'

    r2 = db.atomic(db_path, db.find_nodes({'name': 'Steve'}, db._search_like, db._search_starts_with))
    print(r2)
    assert r2[0]["name"] == 'Steve Wozniak'
    assert r2[1]["name"] == 'Steve Jobs'

    r3 = db.traverse(db_path, 5, neighbors_fn=db.find_inbound_neighbors)
    print(r3)
    assert r3 == [5]

    r4 = db.traverse(db_path, 5, neighbors_fn=db.find_outbound_neighbors)
    print(sorted(r4))
    assert sorted(r4) == [1, 4, 5]

    r5 = db.traverse(db_path, 5, neighbors_fn=db.find_neighbors)
    print(sorted(r5))
    assert sorted(r5) == [1, 2, 3, 4, 5]
