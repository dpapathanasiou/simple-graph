#!/bin/sh


mkdir -p src/simpledb
echo 'package simpledb' > src/simpledb/constants.go
echo '\nconst (' >> src/simpledb/constants.go

for file in $(ls ../sql/*.sql)
do
  sql=$(cat $file)
  val=$(basename $file | sed 's/\.sql//;s/[^-]\+/\L\u&/g;s/-//g')
  echo "    $val = \`$sql" >> src/simpledb/constants.go
  echo '`\n' >> src/simpledb/constants.go
done

echo ')' >> src/simpledb/constants.go
