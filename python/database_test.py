import pytest
import database as db

@pytest.fixture(autouse=True)
def database_test_file(tmp_path):
    d = tmp_path / "simplegraph"
    d.mkdir()
    return d / "apple.sqlite"

@pytest.fixture()
def nodes():
    return {
        1: {'name': 'Apple Computer Company', 'type':['company', 'start-up'], 'founded': 'April 1, 1976'},
        2: {'name': 'Steve Wozniak', 'type':['person','engineer','founder']},
        3: {'name': 'Steve Jobs', 'type':['person','designer','founder']},
        4: {'name': 'Ronald Wayne', 'type':['person','administrator','founder']},
        5: {'name': 'Mike Markkula', 'type':['person','investor']}
    }

def test_initialize(database_test_file):
    db.initialize(database_test_file)
    assert database_test_file.exists()
    assert database_test_file.stat().st_size == 28672

def test_insert_and_search(database_test_file, nodes):
    db.initialize(database_test_file)
    [db.atomic(database_test_file, db.add_node(node, id)) for id, node in nodes.items()]
    for id, node in nodes.items():
        assert db.atomic(database_test_file, db.find_node(id)) == node
