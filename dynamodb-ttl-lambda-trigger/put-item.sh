EXP=`date -d '-1 day' +%s`
RANDOM_ID=`shuf -i1-100 -n1`
aws dynamodb put-item --table-name "ExampleTTL" --item '{"id": {"N": "'$RANDOM_ID'"}, "ttl": {"N": "'$EXP'"}}'
echo "Item succesfully added"