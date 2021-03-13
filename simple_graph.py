import click

from database import initialize, visualize

@click.group()
def cli():
    '''Main entry point '''
    pass

@cli.command()
@click.option('--db_file', default='simple_graph.sqlite', help='SQLite3 back-end.')
@click.option('--schema_file', default='schema.sql', help='Siple_graph schema to use.')
def initialize_database(db_file,schema_file):
    initialize(db_file,schema_file)

@cli.command()
@click.option('--db_file', default='simple_graph.sqlite', help='SQLite3 back-end.')
@click.option('--dot_file', default='simple_graph.dot', help='Graphviz dot output.')
@click.option('--path', default=[], help='Node path criteria.')
@click.option('--frmt', default='png', help='Output file format.')
@click.option('--xn_keys', default=[], help='Exclude node key list.')
@click.option('--hn_keys', default=False, help='Hide node keys?')
@click.option('--node_kv', default=' ', help='Node JSON?')
@click.option('--xe_keys', default=[], help='Exclude edge key list.')
@click.option('--he_keys', default=False, help='Hide edge keys?')
@click.option('--edge_kv', default=' ', help='Edge JSON?')
def visualize_database(db_file,dot_file,path,frmt,xn_keys,hn_keys,node_kv,xe_keys,he_keys,edge_kv):
    visualize(db_file,dot_file,path,frmt,xn_keys,hn_keys,node_kv,xe_keys,he_keys,edge_kv)
    
