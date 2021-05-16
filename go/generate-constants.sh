#!/bin/sh


mkdir -p src/constants
echo 'package constants' > src/constants/constants.go
echo '\nconst (' >> src/constants/constants.go

for file in $(ls ../sql/*.sql)
do
  sql=$(cat $file)
  val=$(basename $file | sed 's/\.sql//;s/[^-]\+/\L\u&/g;s/-//g')
  echo "    $val = \`$sql" >> src/constants/constants.go
  echo '`\n' >> src/constants/constants.go
done

echo ')' >> src/constants/constants.go
