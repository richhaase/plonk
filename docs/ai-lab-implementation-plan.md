
---

### **docs/ai‑lab‑implementation‑plan.md**

```markdown
# AI Lab MVP – Coordinating‑Agent Implementation Plan

## Goals
* One‑command local GenAI stack for solo developers.
* CPU‑first; `gpu` profile optional and explicit.
* No built‑in dataset; weights cached locally via `huggingface-cli`.
* Preserve plonk’s thin orchestrator + Resource architecture.

---

## Phase Breakdown

| Phase | Duration | Key Deliverables | Success Gate |
|-------|----------|------------------|--------------|
| **A. Compose Resource** | 2 days | `dockerComposeResource` + `stack.yml` (CPU & GPU) | `plonk ai-lab up` starts containers; `plonk status -o json` shows Managed. |
| **B. Config & Validation** | 1 day | `model_id`, `model_path`; early‑fail if path invalid | Unit test for bad path; real run passes with cached weights. |
| **C. Lock v2 & Profiles** | 1 day | Lock schema upgrade; GPU profile recorded; v1→v2 migration | Integration test shows lock v2 + profile. |
| **D. Hooks & Dashboard** | 1 day | Pre‑hook Weaviate ping; post‑hook import Grafana dashboard; cost‑meter metrics | Grafana API 200; Prometheus scrape targets up. |
| **E. Manager Simplification** | 1 day | Flatten package managers; helpers.go; delete ErrorMatcher | CI gate: unit+fast tests ≤ 5 s. |
| **F. Docs & Integration Test** | 0.5 day | Overview & README; full end‑to‑end test (CPU) | New dev gets valid answer in <15 s. |

---

## Locked Technical Decisions
* Model weights cached manually; `huggingface-cli` installed via existing PackageManagers.
* Profiles: `default` (CPU) and `gpu` (CUDA).
* Hooks: default timeout 10 m; `continue_on_error` available.
* Reconciliation states: Managed, Missing, Untracked (reserve Degraded for future).
* No GPU auto‑detect or auto‑download in MVP.

---

## Validation Checklist
- [ ] Unit tests green; CI fails if runtime > 5 s.
- [ ] Integration test starts stack, answers “Hello” prompt, writes lock v2.
- [ ] Grafana dashboard shows latency, tokens/s, cost panels.
- [ ] README quick‑start works on clean CPU laptop.
- [ ] Orchestrator ≤ 300 LOC; package LOC targets met.

---

## Post‑MVP Backlog
* Hugging Face auto‑download Resource
* Demo dataset fetcher
* GPU auto‑detect suggestion
* Additional model runtimes (TGI, SGLang)
