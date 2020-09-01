#!/bin/bash
# Cria os resources necessários

echo "Deleting table"
aws dynamodb delete-table --table-name ExampleTTL
