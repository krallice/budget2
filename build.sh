#!/bin/bash

go build main.go
docker build . -t budget2-web
