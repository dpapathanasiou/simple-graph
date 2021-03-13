from functools import partial
from pathlib import Path

import pytest

import database as db


@pytest.fixture
def db_mem_file():
    return ":memory:"


@pytest.fixture(scope="session")
def db_file(tmpdir_factory):
    """Actual file with session scope. """
    fn = tmpdir_factory.mktemp("data").join("test_db.sqlite")
    return fn


@pytest.fixture(scope="session")
def img_file(tmpdir_factory):
    """Place to stick images during tests."""
    fn = tmpdir_factory.mktemp("data").join("apple.png")
    return fn


@pytest.fixture(scope="session")
def curs(db_file):
    """Making a partial function is interesting, but doesn't avoid much typing, as we have to explicitly name the function to call."""
    return partial(db.atomic, db_file=db_file)


@pytest.fixture
def apple_nodes():
    return [
        {
            "name": "Apple Computer Company",
            "type": ["company", "start-up"],
            "founded": "April 1, 1976",
        },
        {"name": "Steve Wozniak", "type": ["person", "engineer", "founder"]},
        {"name": "Steve Jobs", "type": ["person", "designer", "founder"]},
        {"name": "Ronald Wayne", "type": ["person", "administrator", "founder"]},
        {"name": "Mike Markkula", "type": ["person", "investor"]},
    ]


@pytest.fixture()
def n2id(apple_nodes):
    """When our data is a list of dicts with a 'name' property, an enumerated map is handy, so that the ID values are only defined once.
    Might be useful to do

     x[i]=n['name']

    also so that the same dict can be used for outputs to debug.
    """
    x = {}
    for i, n in enumerate(apple_nodes, start=1):
        x[n["name"]] = i
    return x


@pytest.fixture
def apple_connections():
    return [
        ("Steve Wozniak", "Apple Computer Company", {"action": "founded"}),
        ("Steve Jobs", "Apple Computer Company", {"action": "founded"}),
        ("Ronald Wayne", "Apple Computer Company", {"action": "founded"}),
        (
            "Mike Markkula",
            "Apple Computer Company",
            {"action": "invested", "equity": 80000, "debt": 170000},
        ),
        (
            "Apple Computer Company",
            "Ronald Wayne",
            {"action": "divested", "amount": 800, "date": "April 12, 1976"},
        ),
        ("Steve Wozniak", "Steve Jobs", None),
    ]


@pytest.fixture
def apple_upsert_nodes():
    return [("Steve Wozniak", {"nickname": "Woz"})]


@pytest.mark.skip
def test_initialize(db_mem_file):
    """Lack of a file stops us right here"""
    results = db.initialize(db_mem_file)
    assert results == None


@pytest.mark.skip
def test_apple_nodes(db_file, apple_nodes):
    """Running this thwarts TEST_APPLE_GRAPH, since the SQLite file is only created once per session/run."""
    results = db.initialize(db_file)
    assert results == None
    for i, a in enumerate(apple_nodes, start=1):
        db.atomic(db_file, db.add_node(a, i))


def test_apple_graph(db_file, apple_nodes, apple_connections, apple_upsert_nodes, n2id):
    results = db.initialize(db_file)
    assert results == None
    for n in apple_nodes:
        db.atomic(db_file, db.add_node(n, n2id[n["name"]]))
    for c in apple_connections:
        db.atomic(db_file, db.connect_nodes(n2id[c[0]], n2id[c[1]], c[2]))
    for u in apple_upsert_nodes:
        db.atomic(db_file, db.upsert_node(n2id[u[0]], u[1]))


@pytest.mark.skip
def test_partial(
    curs, db_file, apple_nodes, apple_connections, apple_upsert_nodes, n2id
):
    """I didn't like calling db.atomic(db_file,...) every time, so I tried a partial function,
    but now you have to name the second parameter, so it's a tiny aesthetic improvement at best.
    """
    results = db.initialize(db_file)
    assert results == None
    for n in apple_nodes:
        curs(cursor_exec_fn=db.add_node(n, n2id[n["name"]]))
    for c in apple_connections:
        curs(cursor_exec_fn=db.connect_nodes(n2id[c[0]], n2id[c[1]], c[2]))
    for u in apple_upsert_nodes:
        curs(cursor_exec_fn=db.upsert_node(n2id[u[0]], u[1]))


def test_find_via_id(db_file, n2id):
    ret = db.atomic(db_file, db.find_node(n2id["Steve Wozniak"]))
    assert len(ret.items()) == 4


def test_find_via_data(db_file):
    ret = db.atomic(
        db_file,
        db.find_nodes({"name": "Steve"}, db._search_like, db._search_starts_with),
    )
    assert len(ret) == 2


def test_traverse(db_file, n2id):
    ret = db.traverse(db_file, n2id["Steve Wozniak"], n2id["Steve Jobs"])
    assert len(ret) == 2
    ret = db.traverse(db_file, n2id["Ronald Wayne"], n2id["Mike Markkula"])
    assert len(ret) == 3
    ret = db.traverse(
        db_file, n2id["Mike Markkula"], neighbors_fn=db.find_inbound_neighbors
    )
    assert len(ret) == 1
    ret = db.traverse(
        db_file, n2id["Mike Markkula"], neighbors_fn=db.find_outbound_neighbors
    )
    assert len(ret) == 3
    ret = db.traverse(db_file, n2id["Mike Markkula"], neighbors_fn=db.find_neighbors)
    assert len(ret) == 5


def test_visualize(db_file, n2id, img_file):
    p = Path(img_file)
    db.visualize(
        db_file,
        img_file,
        [n2id["Ronald Wayne"], n2id["Apple Computer Company"], n2id["Mike Markkula"]],
    )
    # print(p.stat().st_size)
    assert p.stat().st_size == 404
    db.visualize(
        db_file,
        img_file,
        [n2id["Ronald Wayne"], n2id["Apple Computer Company"], n2id["Mike Markkula"]],
        exclude_node_keys=["type"],
        hide_edge_key=True,
    )
    # print(p.stat().st_size)
    assert p.stat().st_size == 253
