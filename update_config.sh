#!/bin/bash

# Check if the configuration file path is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <config_file_path>"
    exit 1
fi

# Define the configuration file path from the first argument
CONFIG_FILE="$1"

# Check if the configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Configuration file '$CONFIG_FILE' does not exist."
    exit 1
fi

# Get the current year
CURRENT_YEAR=$(date +"%Y")

# Read the download_mode from the configuration file
DOWNLOAD_MODE=$(grep -oP 'download_mode\s*=\s*\K[^ ]+' "$CONFIG_FILE")

# Determine the year to use based on the download_mode
if [ "$DOWNLOAD_MODE" == "Audit" ]; then
    YEAR_TO_USE=$((CURRENT_YEAR - 1))
else
    YEAR_TO_USE=$CURRENT_YEAR
fi

# Replace the old year with the determined year in the configuration file
sed -i "s/download_year = [0-9]\{4\}/download_year = $YEAR_TO_USE/" "$CONFIG_FILE"

# Also update the CheckDirectory if needed
sed -i "s#check_directory = .*/[0-9]\{4\}#check_directory = /home/server/drive/IDX/$YEAR_TO_USE#" "$CONFIG_FILE"

echo "Configuration file '$CONFIG_FILE' updated successfully with the year $YEAR_TO_USE."
