---
steward_assessment: "3"
contract_path: "2026-08-09-repair-bats-homebrew-contract.md"
contract_id: "repair-bats-homebrew-build"
contract_revision: 1
contract_body_sha256: "9d9440219a7e208a48969a603fbe403976938028c40ad218a862ddf71663d1de"
change_identity: "sha256:4af136459259f47939d466a32d7ea60e8697b9ac647a4f1c9911f665eeeabfb3"
environment: "macOS 26.5.2; Homebrew 6.0.15; Docker client 29.7.1; no Docker daemon available"
assessor: "Codex"
created_at: "2026-08-09T18:18:51.420Z"
completed_at: "2026-08-09T18:19:39.799Z"
state: completed
assessment_body_sha256: "6e3f1c54a66bd43cecb9663b2852510c6a8d4c6eb922f6515ebfe2d85c209fb2"
---
# Assessment: Repair BATS container build

## Provenance

- Immutable change: sha256:4af136459259f47939d466a32d7ea60e8697b9ac647a4f1c9911f665eeeabfb3
- Environment/context: macOS 26.5.2; Homebrew 6.0.15; Docker client 29.7.1; no Docker daemon available
- Assessor: Codex

## Overall

Outcome: inconclusive

## Claim outcomes

### AC1: Building the project test container completes without the Homebrew untrusted-tap failure.
- Outcome: inconclusive
- Evidence: E1, E2, E3, E4
- Residual uncertainty: The implementation directly addresses the observed trust failure with documented formula-scoped commands, but no Docker daemon was available to observe a complete image build.

### AC2: The built container provides BATS, bats-support, and bats-assert in the locations expected by the existing integration tests.
- Outcome: inconclusive
- Evidence: E1, E4
- Residual uncertainty: Static inspection shows that the existing installs and `/usr/local/lib` symlinks are preserved, but the resulting filesystem could not be observed without building the image.

### AC3: The container's existing integration-test entry point can execute the BATS suite.
- Outcome: inconclusive
- Evidence: E1, E4
- Residual uncertainty: The entry point and test command are unchanged, but container execution requires an available Docker daemon.

## Evidence log

### E1
- Command or artifact: `git diff --binary -- Dockerfile | shasum -a 256`, producing the assessed change identity `sha256:4af136459259f47939d466a32d7ea60e8697b9ac647a4f1c9911f665eeeabfb3`.
- Observation: The diff adds trust for only `bats-core/bats-core/bats-support` and `bats-core/bats-core/bats-assert`; the existing formula installation, support-library symlinks, and integration entry point are unchanged.

### E2
- Command or artifact: `brew help trust` using Homebrew 6.0.15.
- Observation: The installed Homebrew exposes `brew trust --formula` for trusting named non-official formulae.

### E3
- Command or artifact: Homebrew Tap Trust documentation at `https://docs.brew.sh/Tap-Trust`.
- Observation: Homebrew documents formula-scoped trust before short-name installation and recommends it over trusting an entire third-party tap.

### E4
- Command or artifact: `docker version` and `docker --context colima version`.
- Observation: The Docker 29.7.1 client is installed, but both configured local contexts lack a running daemon socket, preventing image build and container execution.

## Contract observations

- The implementation preserves the contract's trust boundary: it neither trusts the whole BATS tap nor sets Homebrew's global trust opt-out.
- `git diff --check` completed successfully.
- The repository Go suite was exercised separately; all packages except two pre-existing non-TTY color-output tests passed. Those failures do not exercise the Dockerfile change or any acceptance claim.

## Residual risks

- Homebrew formula installation, the resulting symlink targets, and the BATS entry point remain unobserved in a built Linux container.

## Remediation

Classification: insufficient-or-conflicting-evidence
Next action: With a Docker daemon available, run `docker build -t plonk-test:latest .`, verify both support-library symlink targets inside the image, and run `docker run --rm plonk-test:latest all`; then reassess this same frozen contract and immutable change unless the implementation changes.
