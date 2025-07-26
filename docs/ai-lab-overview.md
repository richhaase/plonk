# AI Lab Overview

**Mission**
Turn a brand‑new laptop into a fully‑featured GenAI playground in one command.

## Stack Components

| Service | Image (default profile) | Purpose |
|---------|------------------------|---------|
| **vLLM** (Llama 3 8 B) | `vllm/vllm:latest-cpu` | High‑throughput language‑model server. |
| **Weaviate 2.x** | `semitechnologies/weaviate:2.0.*` | Vector database for RAG workflows. |
| **GuardrailsAI** | `guardrailsai/guardrails:latest` | Output validation / PII stripping. |
| **Cost‑meter** | `plonk/cost-meter:0.1` | Exposes $ / 1 k tokens & GPU stats. |
| **Prometheus** | `prom/prometheus:latest` | Metrics collection. |
| **Grafana** | `grafana/grafana:10` | Live dashboard (latency, tokens/s, cost). |

*CPU images by default; CUDA images selected via `--profile gpu`.*

## Quick‑Start

```bash
# 1‑time: cache model weights (~6 – 12 GB)
huggingface-cli download meta-llama/Meta-Llama-3-8B-Instruct \
    --local-dir ~/models/llama3-8b

# Run AI Lab (CPU)
plonk ai-lab up \
  --model-id meta-llama/Meta-Llama-3-8B-Instruct \
  --model-path ~/models/llama3-8b

# Optional GPU profile
plonk ai-lab up --profile gpu …
```

Visit http://localhost:3000 to see live metrics.

## Model Swapping
Set MODEL_ID & MODEL_PATH in ai-lab.yaml or pass the flags shown above.
Tested alternates: Mistral 7 B v0.3, Gemma 2 9 B, DeepSeek V2.5 7 B.

## Why local cache first?
Local weights keep the MVP small, avoid Hugging Face token handling, and simplify licensing. Auto‑download is a planned enhancement.

## Future Enhancements
GPU auto‑detection suggestion

`plonk ai-lab demo fetch <dataset>` helper

Hugging Face Resource for automatic weight management

Additional runtimes (TGI, SGLang)
