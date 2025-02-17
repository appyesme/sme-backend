#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '#' | awk '/=/ {print $1}')
else
    echo ".env file not found!"
    exit 1
fi

# Prompt user for the table name
read -p "Enter the table names (comma-separated, no space): " table_name

# Check if the user provided a table name
if [ -z "$table_name" ]; then
    echo "Table name cannot be empty!"
    exit 1
fi

# Construct the DSN string
dsn="user=${DB_USERNAME} password=${DB_PASSWORD} host=${DB_HOST} port=${DB_PORT} dbname=${DB_NAME} sslmode=${SSL_MODE}"

# Command to be executed
command="gentool -db postgres -dsn \"${dsn}\" -onlyModel -fieldNullable -fieldWithTypeTag -tables \"${table_name}\""

# Execute the command
eval $command

# Check if the command was successful
if [ $? -eq 0 ]; then
    echo "Command executed successfully."
else
    echo "An error occurred during execution."
fi


# Check if the command was successful
if [ $? -eq 0 ]; then
    echo "Command executed successfully."

    # Move generated model from dao/model to models
    if [ -d "dao/model" ]; then
        mv dao/model/* model/
        echo "Model moved to 'model' folder."

        # Remove the dao/model directory after moving files
        rm -rf dao/
        echo "'dao/' directory has been deleted."
    else
        echo "'dao/model' directory not found."
    fi
else
    echo "An error occurred during execution."
fi
