#!/bin/bash

# Prepare test configuration and, keys
echo "Init .env and, keys" &&
cp .env_circleci .env &&
bash init_tests.sh &&

# Start database
cd arcio-db &&
echo "Starting image"

