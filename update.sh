#!/usr/bin/env bash
podman-compose down
git pull
podman-compose up --build --force -d