#!/bin/bash

keyID="<secret>"
aws-encryption-cli --encrypt --input secret.txt \
                     --master-keys key=$keyID \
                     --encryption-context purpose=test \
                     --metadata-output ~/metadata \
                     --output .