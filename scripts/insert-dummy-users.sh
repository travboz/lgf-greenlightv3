#!/bin/bash

# Define the API endpoint
url="http://localhost:4000/v1/users"

# Define the users to insert
users=(
    '{"name": "Alice Smith", "email": "alice@example.com", "password": "pa55word"}' # succeeds
    '{"name": "Bob Builder", "email": "bob@ibuild.com", "password": "th0ma5555"}' # succeeds
    '{"name": "", "email": "bob@invalid.", "password": "pass"}' # should fail
)

# Loop through each user and insert them using curl
for u in "${users[@]}"; do
    curl -X POST "$url" \
    -H "Content-Type: application/json" \
    -d "$u"
    # echo -e "User created: $u\n"
done