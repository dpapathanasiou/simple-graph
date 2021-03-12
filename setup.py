import setuptools
from pip.req import parse_requirements

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setuptools.setup(
    name="simple-graph", # Replace with your own username
    version="0.0.1",
    author="Denis Papathanasiou",
    author_email="denis@papathanasiou.org",
    description='This is a simple graph database in SQLite, inspired by "SQLite as a document database"',
    install_reqs = parse_requirements('requirements.txt'),
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/dpapathanasiou/simple-graph",
    project_urls={
        "Bug Tracker": "https://github.com/dpapathanasiou/simple-graph/issues",
    },
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    package_dir={"": "."},
    packages=setuptools.find_packages(where="."),
    python_requires=">=3.6",
)
