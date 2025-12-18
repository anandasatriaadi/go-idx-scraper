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
DOWNLOAD_MODE=$(perl -nle 'print $& if /mode\s*:\s*\K[^ ]+/' "$CONFIG_FILE")

echo "Current download mode: $DOWNLOAD_MODE"

# Determine the year to use based on the mode
if [ "$DOWNLOAD_MODE" == "Audit" ]; then
    YEAR_TO_USE=$((CURRENT_YEAR - 1))
else
    YEAR_TO_USE=$CURRENT_YEAR
fi

# Replace the old year with the determined year in the configuration file
# This sed command uses a substitution (s/) to find lines matching "  year: " followed by exactly 4 digits,
# and replaces them with "  year: " followed by the value of YEAR_TO_USE.
# The -i '' flag edits the file in place.
sed -i '' "s/^\(  year: \)[0-9]\{4\}/\1$YEAR_TO_USE/" "$CONFIG_FILE"

# Also update the CheckDirectory if needed
# This sed command uses a substitution to find lines matching "  check_dir: " followed by any characters up to a slash and 4 digits,
# and replaces them with "  check_dir: /home/server/drive/IDX/" followed by YEAR_TO_USE.
# The # is used as the delimiter instead of / to avoid escaping slashes in the path.
# The -i '' flag edits the file in place.
sed -i '' "s#^\(  check_dir: \).*/[0-9]\{4\}#\1/home/server/drive/IDX/$YEAR_TO_USE#" "$CONFIG_FILE"

echo "Configuration file '$CONFIG_FILE' updated successfully with the year $YEAR_TO_USE."
