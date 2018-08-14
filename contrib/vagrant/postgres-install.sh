#!/bin/bash

set -e

# restore database
#cat your_dump.sql | docker exec -i your-db-container psql -Upostgres

docker run -i -d --restart=always \
   -e POSTGRES_USER=lastbackend \
   -e POSTGRES_PASSWORD=lastbackend \
   -e POSTGRES_DB=lastbackend \
   -p 5432:5432 \
   --name=pgs postgres