# This Dockerfile is adapted from the tutorial at https://cloud.google.com/run/docs/tutorials/gpu-gemma2-with-ollama

FROM ollama/ollama:0.3.6

# Listen on all interfaces, port 8080
ENV OLLAMA_HOST 0.0.0.0:8080

# Store model weight files in /models
ENV OLLAMA_MODELS /models

# Reduce logging verbosity
ENV OLLAMA_DEBUG false

# Never unload model weights from the GPU
ENV OLLAMA_KEEP_ALIVE -1 

# Store the model weights in the container image
ENV MODEL yottahmd/tiny-swallow-1.5b-instruct
RUN ollama serve & sleep 5 && ollama pull $MODEL 

# Start Ollama to serve the REST API
ENTRYPOINT ["ollama", "serve"]