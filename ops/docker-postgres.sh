#!/bin/sh

docker volume create pointbotdb

docker run --rm \
    --name pointbot-docker \
    -p 5432:5432 \
    -e POSTGRES_USER=pointbot \
    -e POSTGRES_PASSWORD=pointbot \
    -v pointbotdb:/var/lib/postgresql/data \
    postgres:14-alpine