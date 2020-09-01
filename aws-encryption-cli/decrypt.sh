#!/bin/bash

keyID="<secret>"
aws-encryption-cli --decrypt --input secret.txt.encrypted \
                     --encryption-context purpose=test \
                     --metadata-output ~/metadata \
                     --output .