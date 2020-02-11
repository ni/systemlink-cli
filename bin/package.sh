#! /bin/bash
set -e

# clean up existing zip
rm -f systemlink-cli.zip

# create new zip file
zip -r systemlink-cli.zip build/*
