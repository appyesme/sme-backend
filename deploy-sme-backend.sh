#!/bin/bash

command="go build -o cmd/deploy/sme-backend && cd cmd/deploy/ && sh deploy.sh"
eval $command

if [ $? -eq 0 ]; then
    echo "Command executed successfully."
else
    echo "An error occurred during execution."
fi
