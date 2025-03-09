# GGUF Model Setup for Ollama

This repository contains the necessary files and instructions to convert and publish AI models in GGUF format for use with [Ollama](https://ollama.com). This guide uses the [Tiny Swallow 1.5B model](https://sakana.ai/taid-jp/) from Sakana AI as an example, but the process is applicable to other models.

## Overview

[GGUF](https://github.com/ggml-org/ggml/blob/master/docs/gguf.md) is a file format for storing models for inference that can be used With Ollama. This repository helps you:

1. Convert models to GGUF format (or use pre-converted models)
2. Create Modelfiles for Ollama
3. Publish and use models locally with Ollama

## Prerequisites
- [uv (latest stable)](https://docs.astral.sh/uv/getting-started/installation/)
- [llama.cpp (latest stable)](https://github.com/ggml-org/llama.cpp/blob/master/docs/install.md)
- [Ollama (latest stable)](https://ollama.com/download)

## Steps to convert HF model to GGUF

Note: This instruction is largely inspired by [this tutorial](https://github.com/ggml-org/llama.cpp/discussions/2948)

1. Download the model from Hugging Face 
  ```bash
  ./download.sh
  ```

2. Convert the model to GGUF
  ```bash
  git submodule update --init --recommend-shallow
  uv add -r llama.cpp/requirements.txt
  uv run llama.cpp/convert_hf_to_gguf.py ./models/tiny-swallow-1.5b-instruct --outfile out/tiny-swallow-1.5b-instruct.gguf
  ```

## Publish the model on Ollama

1. Create a new repository <username>/<my-model> on [Ollama](https://ollama.com)

2. Create the `Modelfile` with the following content
  ```yaml
  FROM </path/to/model>.gguf

3. Push the model to the repository
  ```bash
  ollama create <username>/<my-model>
  ollama push <username>/<my-model>
  ```