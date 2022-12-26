#!/bin/sh

# this script produces the sql constants for the go module,
# based on the contents of the /sql folder, one for each file

mkdir -p simplegraph
target=$(echo "simplegraph/constants.go")

echo 'package simplegraph' > $target
echo '\nconst (' >> $target

# sql files: ready for bindings as-is
for file in $(ls ../sql/*.sql)
do
  sql=$(cat $file)
  val=$(basename $file | sed 's/\.sql//;s/[^-]\+/\L\u&/g;s/-//g')
  echo "    $val = \`$sql" >> $target
  echo '`\n' >> $target
done

# template files: need updating to go syntax (https://pkg.go.dev/text/template)
for file in $(ls ../sql/*.template)
do
  sql=$(cat $file | sed 's/{% endif %}/{{ end }}/g;s/{% endfor %}/{{ end }}/g;s/{% if/{{ if/g;s/{% for [a-zA-Z_]\+ in /{{ range /g;s/ %}/ }}/g;s/[^_]\+ }}/\L\u& }}/g;s/_//g;s/[a-zA-Z_]\+ }}/\.\u&/g;s/ }} }}/ }}/g;s/{{ .End }}/{{ end }}/g')
  val=$(basename $file | sed 's/\.template/-template/;s/[^-]\+/\L\u&/g;s/-//g')
  echo "    $val = \`$sql" >> $target
  echo '`\n' >> $target
done

echo ')' >> $target
