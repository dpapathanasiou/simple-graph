from setuptools import find_packages, setup

with open('simple_graph_db/__init__.py', 'r') as f:
    for line in f:
        if line.startswith('__version__'):
            version = line.strip().split('=')[1].strip(' \'"')
            break
    else:
        version = '0.0.1'

with open('README.md', 'rb') as f:
    readme = f.read().decode('utf-8')

REQUIRES = []

setup(
    name='simple_graph_db',
    version=version,
    description='This is a simple graph database in SQLite, inspired by "SQLite as a document database".',
    long_description=readme,
    long_description_content_type="text/markdown",
    author="Denis Papathanasiou",
    author_email="denis@papathanasiou.org",
    maintainer='dpapathanasiou',
    maintainer_email='denis@papathanasiou.org',
    url="https://github.com/dpapathanasiou/simple-graph",
    license='MIT',

    keywords=[
        'graph database',
    ],

    classifiers=[
        'Development Status :: 3 - Alpha',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: MIT License',
        'Natural Language :: English',
        'Operating System :: OS Independent',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
        'Programming Language :: Python :: Implementation :: CPython',
        'Programming Language :: Python :: Implementation :: PyPy',
    ],

    install_requires=REQUIRES,
    include_package_data=True,
    tests_require=['coverage', 'pytest'],
    test_suite='tests',
    packages=find_packages(exclude=["tests"]),
)
