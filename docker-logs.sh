#!/bin/bash

services=$(docker compose config --services)
for service in $services; do
    docker compose logs -f --tail=200 $service &
done
wait