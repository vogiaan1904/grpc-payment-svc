#!/bin/bash

# Script to copy proto files from the local protos-submodule to protos directory
# This allows order-svc to have its own copy of proto files for code generation

# Define paths
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_DIR="$ROOT_DIR/protos-submodule/proto"
DEST_DIR="$ROOT_DIR/protos"

main() {
  echo "Updating protos submodule..."
  # Navigate to project root and update the protos submodule
  cd "$ROOT_DIR"
  git submodule update --remote protos-submodule
  
  # Check if source directory exists
  if [ ! -d "$SOURCE_DIR" ]; then
    echo "Error: Source directory $SOURCE_DIR does not exist."
    echo "Make sure the protos-submodule directory exists."
    exit 1
  fi

  # Create destination directory if it doesn't exist
  mkdir -p "$DEST_DIR"

  # Copy proto files
  cp -f "$SOURCE_DIR"/*.proto "$DEST_DIR/"
  
  echo "Proto files copied to $DEST_DIR"
}

main 