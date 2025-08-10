#!/bin/bash

# Define the API endpoint
url="http://localhost:4000/v1/movies"

# Define the users to insert
movies=(
    '{"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["action","adventure"]}'
    '{"title":"Deadpool","year":2016, "runtime":"108 mins","genres":["action","comedy"]}'
    '{"title":"The Breakfast Club","year":1986, "runtime":"96 mins","genres":["drama"]}'
    '{"title":"Moana","year":2016,"runtime":"107 mins", "genres":["animation","adventure"]}'
    '{"title":"Toy Story","year":1995,"runtime":"81 mins","genres":["animation","adventure","kids"]}'
    '{"title":"Finding Nemo","year":2003,"runtime":"100 mins","genres":["animation","adventure","kids"]}'
    '{"title":"Frozen","year":2013,"runtime":"102 mins","genres":["animation","adventure","kids"]}'
)

# Loop through each user and insert them using curl
for movie in "${movies[@]}"; do
    curl -X POST "$url" \
    -H "Content-Type: application/json" \
    -d "$movie"
    # echo -e "Movie inserted: $movie\n"
done