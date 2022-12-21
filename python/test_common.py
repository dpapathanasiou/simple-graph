import pytest
import database as db


@pytest.fixture()
def database_test_file(tmp_path):
    d = tmp_path / "simplegraph"
    d.mkdir()
    return d / "apple.sqlite"


@pytest.fixture()
def nodes():
    return {
        1: {'name': 'Apple Computer Company', 'type': ['company', 'start-up'], 'founded': 'April 1, 1976'},
        2: {'name': 'Steve Wozniak', 'type': ['person', 'engineer', 'founder']},
        '3': {'name': 'Steve Jobs', 'type': ['person', 'designer', 'founder']},
        4: {'name': 'Ronald Wayne', 'type': ['person', 'administrator', 'founder']},
        5: {'name': 'Mike Markkula', 'type': ['person', 'investor']}
    }


@pytest.fixture()
def edges():
    return {
        1: [(4, {'action': 'divested', 'amount': 800, 'date': 'April 12, 1976'})],
        2: [(1, {'action': 'founded'}), ('3', None)],
        '3': [(1, {'action': 'founded'})],
        4: [(1, {'action': 'founded'})],
        5: [(1, {'action': 'invested', 'equity': 80000, 'debt': 170000})]
    }


@pytest.fixture()
def apple(database_test_file, nodes, edges):
    db.initialize(database_test_file)
    [db.atomic(database_test_file, db.add_node(node, id))
     for id, node in nodes.items()]
    for src, targets in edges.items():
        for target in targets:
            tgt, label = target
            if label:
                db.atomic(database_test_file,
                          db.connect_nodes(src, tgt, label))
            else:
                db.atomic(database_test_file, db.connect_nodes(src, tgt))
    yield
