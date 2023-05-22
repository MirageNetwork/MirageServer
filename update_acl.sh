#!/bin/bash

# Check that one or two arguments are provided
if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "Usage: $0 read [id]|validate|write id|list"
  exit 1
fi

# Get the command argument
command="$1"

if [ "$command" = "read" ]; then
  # Get the id argument if provided
  id="$2"
  if [ -z "$id" ]; then
    id="1648881622080032768"
  fi

  # Use sqlite3 to query the organizations table
  ACL=$(sqlite3 db.sqlite "SELECT acl_policy FROM organizations WHERE id = '${id}';")

  # Use jq to process the data and write it to a file called r.txt
  echo "${ACL}" | jq '.' > r.txt
  echo "Data written to r.txt"
elif [ "$command" = "validate" ]; then
  # Validate that r.txt contains valid JSON using jq
  if jq empty < "r.txt" > /dev/null 2>&1; then
    echo "r.txt contains valid JSON"
  else
    echo "r.txt does not contain valid JSON"
    exit 1
  fi
elif [ "$command" = "write" ]; then
  # Check that the id argument is provided
  if [ -z "$2" ]; then
    echo "Usage: $0 write id"
    exit 1
  fi

  # Get the id argument
  id="$2"

  # Use jq to process the data in r.txt
  acl_policy=$(jq '.' < r.txt)

  # Validate that r.txt contains valid JSON using jq
  if ! jq empty < "r.txt" > /dev/null 2>&1; then
    echo "r.txt does not contain valid JSON"
    exit 1
  fi

  # Use sqlite3 to update the organizations table
  sqlite3 db.sqlite <<EOF
  UPDATE organizations
  SET acl_policy = '${acl_policy}'
  WHERE id = '${id}';
EOF
  echo "Data written to database for id ${id}"
elif [ "$command" = "list" ]; then
  # Use sqlite3 to query the organizations table
  sqlite3 -header -column db.sqlite "SELECT id, name FROM organizations;"
else
  echo "Usage: $0 read [id]|validate|write id|list"
  exit 1
fi

