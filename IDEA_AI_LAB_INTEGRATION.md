# IDEA: AI Lab Integration & Future Resource Management

## Current State

- Planned commands: `plonk ai-lab up` and `plonk ai-lab down`
- Will manage Docker Compose stacks for AI services
- Uses same Resource interface as packages/dotfiles
- Future: might add more resource types (databases, queues, etc.)

## Issues to Consider

1. **Command Namespace**: Is `ai-lab` too specific?
2. **Pattern Consistency**: How do other resources get managed?
3. **Verb Choice**: up/down vs start/stop vs enable/disable
4. **Resource Discovery**: How do users know what resources are available?

## Potential Solutions

### Option A: Specific Subcommands (Current Plan)
```bash
plonk ai-lab up
plonk ai-lab down
plonk ai-lab status
# Future:
plonk database up
plonk queue up
```

**Pros**:
- Clear what each command does
- Can have resource-specific options
- Follows docker-compose pattern

**Cons**:
- New pattern for each resource type
- Inconsistent with package/dotfile management

### Option B: Generic Resource Commands
```bash
plonk up ai-lab
plonk down ai-lab
plonk status ai-lab
# Or:
plonk start ai-lab
plonk stop ai-lab
```

**Pros**:
- Extensible pattern
- Consistent verbs
- Could work with all resources

**Cons**:
- Less specific
- Might not fit all resource types

### Option C: Resource-Aware Sync
```bash
# Everything through sync/apply
plonk sync                    # reconciles all resources
plonk sync --only packages    # just packages
plonk sync --only ai-lab      # just AI lab
plonk apply                   # maybe better name than sync?
```

**Pros**:
- Single command to learn
- Declarative model
- Config-driven

**Cons**:
- Less direct control
- How to stop services?

### Option D: Service-Oriented Commands
```bash
plonk services list
plonk services enable ai-lab
plonk services disable ai-lab
plonk services status
```

**Pros**:
- Clear mental model
- Extensible to any service
- Similar to systemd patterns

**Cons**:
- Another command namespace
- Not all resources are "services"

### Option E: Hybrid Approach
```bash
# Direct control for services
plonk up ai-lab              # start service
plonk down ai-lab            # stop service

# Declarative for packages/dotfiles
plonk sync                   # reconcile all
plonk add package ripgrep    # modify config and sync
```

**Pros**:
- Best of both worlds
- Services need different verbs than packages
- Clear mental model

**Cons**:
- Inconsistent patterns
- More to learn

## Questions for Discussion

1. Should all resources follow the same command pattern?
2. Is `ai-lab` the right name, or should it be more generic?
3. How do we handle resources that need runtime control (start/stop)?
4. Should everything go through config + sync?

## Command Patterns in Other Tools

**Docker Compose:**
```bash
docker-compose up
docker-compose down
docker-compose ps
```

**Kubernetes:**
```bash
kubectl apply -f config.yaml
kubectl delete -f config.yaml
kubectl get pods
```

**Systemd:**
```bash
systemctl start service
systemctl stop service
systemctl enable service
```

## Future Resource Types to Consider

1. **Docker Compose Stacks**: Need up/down/logs
2. **Databases**: Need init/start/stop/backup?
3. **Model Weights**: Need download/cache/clean?
4. **Development Environments**: Need create/destroy/enter?

## Recommendation Placeholder

_To be filled after discussion_
