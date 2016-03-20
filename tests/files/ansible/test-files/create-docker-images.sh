#!/bin/bash -x

# creates the necessary docker images to run testrunner.sh locally

docker build --tag="shift/cppjit-testrunner" docker-cppjit
docker build --tag="shift/python-testrunner" docker-python
docker build --tag="shift/go-testrunner" docker-go
