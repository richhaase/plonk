# Container/VM Testing Strategy

## Current Reality

1. **--dry-run exists but insufficient**: Shows what would happen but still requires full system setup
2. **Architecture prevents mocking**: Tight coupling means we can't mock package managers
3. **System modification is inherent**: Plonk's purpose is to modify systems

## Container/VM Testing Approach

### 1. Docker-Based Testing (Linux Only)

**Advantages:**
- Fast to spin up/down
- Good for Linux testing
- Easy CI integration
- Can test multiple distros

**Limitations:**
- No macOS support (plonk's primary platform)
- Can't test Homebrew on Linux properly
- Missing macOS-specific behaviors

**Implementation:**
```dockerfile
# Dockerfile.test.ubuntu
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y curl git golang
COPY . /plonk
WORKDIR /plonk
RUN go build ./cmd/plonk
CMD ["make", "test-integration"]
```

### 2. Lima VM Testing (Current Approach)

**Advantages:**
- Works on macOS
- Can test both macOS and Linux
- Already partially implemented
- Closer to real user environment

**Current Usage:**
```bash
# Manual process
lima create --name=plonk-test
lima shell plonk-test
# Run tests manually
```

**Proposed Automation:**
```bash
#!/bin/bash
# scripts/test-in-lima.sh
VM_NAME="plonk-test-$$"
lima create --name="$VM_NAME" --cpus=2 --memory=4 ./tests/lima/test-vm.yaml
lima copy -r . "$VM_NAME:/tmp/plonk"
lima shell "$VM_NAME" -- bash -c "cd /tmp/plonk && make test-integration"
lima delete "$VM_NAME"
```

### 3. GitHub Actions VMs (CI Only)

**Current Setup:**
- Uses GitHub's macOS and Ubuntu runners
- Already isolated environments
- But still installs packages on runner

**Improvement:**
- Create ephemeral test users
- Use separate package manager prefixes
- Clean environment between tests

### 4. Vagrant Option

**Advantages:**
- Cross-platform VM management
- Supports macOS guests (with effort)
- Reproducible environments

**Disadvantages:**
- Heavier than containers
- Requires VirtualBox/VMware
- Slower than Lima

## Recommended Implementation Plan

### Phase 1: Automated Lima Testing (Immediate)

1. Create Lima VM templates:
   ```yaml
   # tests/lima/ubuntu-test.yaml
   images:
   - location: "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
   mounts:
   - location: "."
     mountPoint: "/workspace"
   provision:
   - mode: system
     script: |
       apt-get update
       apt-get install -y golang git curl
   ```

2. Add Makefile targets:
   ```makefile
   test-lima-ubuntu:
       ./scripts/test-in-lima.sh ubuntu-test.yaml

   test-lima-all: test-lima-ubuntu test-lima-debian
   ```

3. Document for developers:
   ```bash
   # Safe local testing
   make test-unit        # Pure functions only
   make test-lima-ubuntu # Full tests in VM
   ```

### Phase 2: Docker for Linux Testing

1. Create test containers for each distro
2. Run BATS tests inside containers
3. Parallel execution for speed

### Phase 3: Enhance BATS Tests

1. Add VM detection:
   ```bash
   if [[ ! -f /.dockerenv ]] && [[ ! -f /dev/vmm ]]; then
     echo "Tests must run in container or VM"
     exit 1
   fi
   ```

2. Make tests more aggressive in VMs:
   - Test error cases
   - Test system limits
   - Test concurrent operations

## Developer Workflow

### Safe Testing Levels

1. **Level 0: Unit Tests** (always safe)
   ```bash
   go test ./...
   ```

2. **Level 1: Dry Run Tests** (mostly safe)
   ```bash
   PLONK_TEST_DRY_RUN=true make test-integration
   ```

3. **Level 2: Container Tests** (safe, Linux only)
   ```bash
   make test-docker
   ```

4. **Level 3: VM Tests** (safe, all platforms)
   ```bash
   make test-lima
   ```

5. **Level 4: Real System Tests** (unsafe, CI only)
   ```bash
   # Only in GitHub Actions
   make test-system
   ```

## Key Decisions

### What Stays in BATS
- User workflow tests
- CLI behavior verification
- Output format testing
- Multi-command scenarios

### What Moves to Go
- Unit tests for business logic
- Component integration tests
- Error injection tests
- Performance tests

### VM/Container Strategy
- **Local Development**: Lima VMs or Docker
- **CI**: GitHub Actions native + containers
- **Release Testing**: Full VM matrix

## Implementation Priority

1. **Immediate (for v1.0)**:
   - Automate Lima VM testing
   - Add safety checks to BATS
   - Document safe testing practices

2. **Short Term**:
   - Docker test containers
   - Parallel test execution
   - Better test isolation

3. **Long Term**:
   - Full VM automation
   - Cross-platform test matrix
   - Performance benchmarks

## Conclusion

Since we can't mock effectively and --dry-run isn't sufficient, VM/container testing is the only safe approach. Lima provides the best balance of safety and platform coverage for local development, while Docker offers speed for Linux-specific testing.
