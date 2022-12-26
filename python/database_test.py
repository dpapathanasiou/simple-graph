import sqlite3
import json
from stat import S_ISREG
import database as db
from test_common import database_test_file, nodes, edges, apple


def test_initialize(database_test_file, apple):
    assert database_test_file.exists()
    assert S_ISREG(database_test_file.stat().st_mode)


def test_bulk_operations(database_test_file, nodes, edges):
    db.initialize(database_test_file)
    ids = []
    bodies = []
    for id, body in nodes.items():
        ids.append(id)
        bodies.append(body)

    # bulk add and confirm
    db.atomic(database_test_file, db.add_nodes(bodies, ids))
    for id, node in nodes.items():
        assert db.atomic(database_test_file, db.find_node(id)) == node

    # bulk upsert and confirm
    db.atomic(database_test_file, db.upsert_nodes(bodies, ids))
    for id, node in nodes.items():
        assert db.atomic(database_test_file, db.find_node(id)) == node

    # bulk connect and confirm
    sources = []
    targets = []
    properties = []
    for src, tgts in edges.items():
        for target in tgts:
            tgt, label = target
            sources.append(src)
            targets.append(tgt)
            if label:
                properties.append(label)
            else:
                properties.append({})

    db.atomic(database_test_file, db.connect_many_nodes(
        sources, targets, properties))
    for src, tgts in edges.items():
        actual = [tuple(x) for x in [[edge[0], edge[1], json.loads(edge[2])]
                                     for edge in db.atomic(database_test_file, db.get_connections_one_way(src))]]
        for target in tgts:
            tgt, label = target
            if label:
                expected = (str(src), str(tgt), label)
            else:
                expected = (str(src), str(tgt), {})
            assert expected in actual

    # bulk remove and confirm
    db.atomic(database_test_file, db.remove_nodes(ids))
    for id in ids:
        assert db.atomic(database_test_file, db.find_node(id)) == {}


def test_exception(database_test_file, apple, nodes):
    node_id = 1
    try:
        db.atomic(database_test_file, db.add_node(nodes[node_id], node_id))
    except sqlite3.IntegrityError as e:  # should be thrown since we are inserting a duplicate node
        assert 'UNIQUE constraint failed: nodes.id' in e.args

    # Inserting a new node should work
    new_node = {'name': 'Tim Cook', 'type': ['person', 'CEO']}
    new_node_id = 6
    try:
        db.atomic(database_test_file, db.add_node(new_node, new_node_id))
    except Exception as e:
        assert False


def test_search(database_test_file, apple, nodes):
    # search by id
    for id, node in nodes.items():
        assert db.atomic(database_test_file, db.find_node(id)) == node

    # simple key-value search using 'LIKE'
    kv_name_like = db._generate_clause('name', predicate='LIKE')
    steves = db.atomic(database_test_file,
                       db.find_nodes([kv_name_like],
                                     ('Steve%',)))
    assert len(steves) == 2
    assert list(map(lambda x: x['name'], steves)) == ['Steve Wozniak',
                                                      'Steve Jobs']

    # plus an additional binding on the same key to find a different person
    kv_other_name = db._generate_clause('name', predicate='LIKE', joiner='OR')
    woz_mike = db.atomic(database_test_file,
                         db.find_nodes([kv_name_like, kv_other_name],
                                       ('%Woz%', '%Markkula',)))
    assert len(woz_mike) == 2
    assert list(map(lambda x: x['name'], woz_mike)) == ['Steve Wozniak',
                                                        'Mike Markkula']

    # a tree search to find all the people with 'founder' in their 'type' list
    founders = db.atomic(database_test_file,
                         db.find_nodes([db._generate_clause('type', tree=True)],
                                       ('founder',),
                                       tree_query=True,
                                       key='type'))
    assert len(founders) == 3
    assert list(map(lambda x: x['name'], founders)) == ['Steve Wozniak',
                                                        'Steve Jobs',
                                                        'Ronald Wayne']


def test_traversal(database_test_file, apple):
    # the traversal CTE seed type is respected, and appears in the output as-is
    assert db.traverse(database_test_file, 2, 3) == [2, '1', '3']
    # singly-quoted strings works as expected
    assert db.traverse(database_test_file, '2', '3') == [
        '2', '1', '3', '4', '5']
    # and so do doubly-quoted values (in the prior version, these produced empty lists)
    assert db.traverse(database_test_file, "2", "3") == [
        '2', '1', '3', '4', '5']

    # more test sets of this pattern:
    # since int is a different type than string, it can appear twice in the output
    assert db.traverse(database_test_file, 4, 5) == [
        4, '1', '2', '3', '4', '5']
    assert db.traverse(database_test_file, '4', '5') == [
        '4', '1', '2', '3', '5']
    assert db.traverse(database_test_file, "4", "5") == [
        '4', '1', '2', '3', '5']

    assert db.traverse(database_test_file, 5,
                       neighbors_fn=db.find_inbound_neighbors) == [5]
    assert db.traverse(database_test_file, '5',
                       neighbors_fn=db.find_inbound_neighbors) == ['5']
    assert db.traverse(database_test_file, 5,
                       neighbors_fn=db.find_outbound_neighbors) == [5, '1', '4']
    assert db.traverse(database_test_file, '5',
                       neighbors_fn=db.find_outbound_neighbors) == ['5', '1', '4']
    assert db.traverse(database_test_file, 5, neighbors_fn=db.find_neighbors) == [
        5, '1', '2', '3', '4', '5']
    assert db.traverse(database_test_file, '5', neighbors_fn=db.find_neighbors) == [
        '5', '1', '2', '3', '4']
    assert db.traverse(database_test_file, "5", neighbors_fn=db.find_neighbors) == [
        '5', '1', '2', '3', '4']


def test_traversal_with_bodies(database_test_file, apple):
    def _normalize_results(results):
        return [(x, y, json.loads(z)) for (x, y, z) in results]

    assert _normalize_results(db.traverse_with_bodies(database_test_file, 2, 3)) == _normalize_results(
        [('2', '()', '{"name":"Steve Wozniak","type":["person","engineer","founder"],"id":2}'),
         ('1', '->', '{"action":"founded"}'), ('3', '->', '{}'), (
         '1', '()', '{"name":"Apple Computer Company","type":["company","start-up"],"founded":"April 1, 1976","id":1}'),
         (
             '2', '<-', '{"action":"founded"}'), ('3', '<-', '{"action":"founded"}'),
         ('4', '<-', '{"action":"founded"}'), ('5', '<-',
                                               '{"action":"invested","equity":80000,"debt":170000}'),
         ('4', '->', '{"action":"divested","amount":800,"date":"April 12, 1976"}'),
         ('3', '()', '{"name":"Steve Jobs","type":["person","designer","founder"],"id":"3"}')])
    assert _normalize_results(
        db.traverse_with_bodies(database_test_file, 5, neighbors_fn=db.find_inbound_neighbors)) == _normalize_results(
        [('5', '()', '{"name":"Mike Markkula","type":["person","investor"],"id":5}')])
    assert _normalize_results(
        db.traverse_with_bodies(database_test_file, 5, neighbors_fn=db.find_outbound_neighbors)) == _normalize_results(
        [('5', '()', '{"name":"Mike Markkula","type":["person","investor"],"id":5}'),
         ('1', '->', '{"action":"invested","equity":80000,"debt":170000}'), (
             '1', '()',
             '{"name":"Apple Computer Company","type":["company","start-up"],"founded":"April 1, 1976","id":1}'),
         ('4', '->', '{"action":"divested","amount":800,"date":"April 12, 1976"}'),
         ('4', '()',
          '{"name":"Ronald Wayne","type":["person","administrator","founder"],"id":4}'),
         ('1', '->', '{"action":"founded"}')])
    assert _normalize_results(
        db.traverse_with_bodies(database_test_file, 5, neighbors_fn=db.find_neighbors)) == _normalize_results(
        [('5', '()', '{"name":"Mike Markkula","type":["person","investor"],"id":5}'),
         ('1', '->', '{"action":"invested","equity":80000,"debt":170000}'), (
         '1', '()', '{"name":"Apple Computer Company","type":["company","start-up"],"founded":"April 1, 1976","id":1}'),
         ('2', '<-', '{"action":"founded"}'), ('3', '<-',
                                               '{"action":"founded"}'), ('4', '<-', '{"action":"founded"}'),
         (
             '5', '<-', '{"action":"invested","equity":80000,"debt":170000}'),
         ('4', '->', '{"action":"divested","amount":800,"date":"April 12, 1976"}'),
         ('2', '()',
          '{"name":"Steve Wozniak","type":["person","engineer","founder"],"id":2}'),
         ('1', '->', '{"action":"founded"}'), ('3', '->', '{}'),
         ('3', '()',
          '{"name":"Steve Jobs","type":["person","designer","founder"],"id":"3"}'), ('2', '<-', '{}'),
         ('4', '()',
          '{"name":"Ronald Wayne","type":["person","administrator","founder"],"id":4}'),
         ('1', '<-', '{"action":"divested","amount":800,"date":"April 12, 1976"}')])
