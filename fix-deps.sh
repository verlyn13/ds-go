#!/bin/bash
# Fix Go module dependencies

cd /home/verlyn13/Projects/ds-go

echo "Fixing Go module dependencies..."

# Download and verify modules
go mod tidy

# Initialize git repo (needed for version in Makefile)
git init
git add .
git commit -m "Initial commit"

echo "Dependencies fixed!"
echo ""
echo "Now you can run:"
echo "  make build"
echo "  ./ds status"