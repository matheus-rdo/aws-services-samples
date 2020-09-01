#!/bin/bash
# Cria os resources necessários

echo "Creating DynamoDB Table"
aws dynamodb create-table \
    --table-name ExampleTTL \
    --attribute-definitions AttributeName=id,AttributeType=N \
    --key-schema AttributeName=id,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST

echo "Waiting for AWS to create the table ..."

sleep 10

echo "Enabling TTL"
aws dynamodb update-time-to-live --table-name ExampleTTL --time-to-live-specification "Enabled=true, AttributeName=ttl"

echo "Resources successfully created"