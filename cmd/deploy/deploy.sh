#!/bin/bash

# Load environment variables from .env
if [ -f ".env" ]; then
    export $(cat .env | xargs)
else
    echo ".env not found."
    exit 1
fi

# Check if required environment variables are set
if [[ -z "$SME_SSH_USER" || -z "$SME_SSH_HOST" || -z "$SME_SSH_PASSWORD" ]]; then
    echo "Missing required environment variables. Ensure SME_SSH_USER, SME_SSH_HOST, and SME_SSH_PASSWORD are set."
    exit 1
fi

# Commands for deployment
# Deployment -- START --
sshpass -p "$SME_SSH_PASSWORD" ssh "$SME_SSH_USER@$SME_SSH_HOST" "mkdir -p /root/apps/sme/sme-new"
sshpass -p "$SME_SSH_PASSWORD" scp sme-backend "$SME_SSH_USER@$SME_SSH_HOST:/root/apps/sme/sme-new"
sshpass -p "$SME_SSH_PASSWORD" ssh "$SME_SSH_USER@$SME_SSH_HOST" "sudo systemctl stop sme.service"
sshpass -p "$SME_SSH_PASSWORD" ssh "$SME_SSH_USER@$SME_SSH_HOST" "if [ -f /root/apps/sme/sme-backend ]; then rm /root/apps/sme/sme-backend; fi"
sshpass -p "$SME_SSH_PASSWORD" ssh "$SME_SSH_USER@$SME_SSH_HOST" "mv /root/apps/sme/sme-new/sme-backend /root/apps/sme/sme-backend"
sshpass -p "$SME_SSH_PASSWORD" ssh "$SME_SSH_USER@$SME_SSH_HOST" "sudo systemctl restart sme.service"
# Deployment -- END --

# Deleting 'sme-backend' binary file.
# Cleanup: Delete the local 'sme-backend' file if the deployment was successful
if [ $? -eq 0 ]; then
	echo "Deployment completed successfully."
    echo "Cleaning up local 'sme-backend' file..."
    rm -f sme-backend
    if [ $? -eq 0 ]; then
        echo "'sme-backend' file deleted successfully."
    else
        echo "Failed to delete 'sme-backend' file."
    fi
else
    echo "Deployment failed. Skipping cleanup."
    exit 1
fi
