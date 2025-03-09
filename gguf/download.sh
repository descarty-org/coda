#!/bin/bash
set -e

# Create and activate virtual environment
uv venv .venv
source .venv/bin/activate

# Install dependencies from pyproject.toml
uv sync

# Download the model files
huggingface-cli download SakanaAI/TinySwallow-1.5B-Instruct \
  --local-dir ./models/tiny-swallow-1.5b-instruct \
  --local-dir-use-symlinks False
