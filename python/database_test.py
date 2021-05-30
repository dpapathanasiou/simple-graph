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
        3: {'name': 'Steve Jobs', 'type': ['person', 'designer', 'founder']},
        4: {'name': 'Ronald Wayne', 'type': ['person', 'administrator', 'founder']},
        5: {'name': 'Mike Markkula', 'type': ['person', 'investor']}
    }


@pytest.fixture()
def edges():
    return {
        1: [(4, {'action': 'divested', 'amount': 800, 'date': 'April 12, 1976'})],
        2: [(1, {'action': 'founded'}), (3, None)],
        3: [(1, {'action': 'founded'})],
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


def test_initialize(database_test_file, apple):
    assert database_test_file.exists()
    assert database_test_file.stat().st_size == 28672


def test_search(database_test_file, apple, nodes):
    for id, node in nodes.items():
        assert db.atomic(database_test_file, db.find_node(id)) == node
    steves = db.atomic(database_test_file, db.find_nodes(
        {'name': 'Steve'}, db._search_like, db._search_starts_with))
    assert len(steves) == 2
    assert list(map(lambda x: x['name'], steves)) == [
        'Steve Wozniak', 'Steve Jobs']


def test_traversal(database_test_file, apple):
    assert db.traverse(database_test_file, 2, 3) == ['2', '1', '3']
    assert db.traverse(database_test_file, 4, 5) == ['4', '1', '2', '3', '5']
    assert db.traverse(database_test_file, 5,
                       neighbors_fn=db.find_inbound_neighbors) == ['5']
    assert db.traverse(database_test_file, 5,
                       neighbors_fn=db.find_outbound_neighbors) == ['5', '1', '4']
    assert db.traverse(database_test_file, 5, neighbors_fn=db.find_neighbors) == [
        '5', '1', '2', '3', '4']
