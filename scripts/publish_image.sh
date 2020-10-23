#!/bin/bash
name=$1
docker build -f build/Dockerfile_${name} . -t docker.io/alitari/ce-go-template-${name}
docker push docker.io/alitari/ce-go-template-${name}