#!/usr/bin/bash
set -ex
VERSION=$(cat VERSION.txt)
git tag v$VERSION
git push --tags
