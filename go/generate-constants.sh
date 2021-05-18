#!/bin/sh


mkdir -p simplegraph
echo 'package simplegraph' > simplegraph/constants.go
echo '\nconst (' >> simplegraph/constants.go

for file in $(ls ../sql/*.sql)
do
  sql=$(cat $file)
  val=$(basename $file | sed 's/\.sql//;s/[^-]\+/\L\u&/g;s/-//g')
  echo "    $val = \`$sql" >> simplegraph/constants.go
  echo '`\n' >> simplegraph/constants.go
done

echo ')' >> simplegraph/constants.go
