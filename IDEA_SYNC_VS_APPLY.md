# IDEA: Sync vs Apply - Core Command Naming

## Current State

- Current command: `plonk sync`
- This is the core reconciliation command
- User note: "I like apply, but in an earlier session an AI agent convinced me to change it"

## Background

The command that reads config and reconciles system state is central to plonk's operation. The name sets user expectations about what will happen.

## Issues to Consider

1. **Mental Model**: What does the user expect to happen?
2. **Industry Patterns**: What do similar tools use?
3. **Directionality**: Does it imply one-way or two-way?
4. **Destructiveness**: Does it sound safe or dangerous?

## Command Name Options

### Option A: `sync` (Current)
```bash
plonk sync
```

**Pros**:
- Implies bringing things into alignment
- Familiar from rsync, browser sync
- Sounds bidirectional

**Cons**:
- Might imply two-way sync (it's not)
- Less clear about source of truth
- Could mean "download" to some users

### Option B: `apply`
```bash
plonk apply
```

**Pros**:
- Clear one-way operation
- Config is applied to system
- Follows Kubernetes pattern
- Declarative feel

**Cons**:
- Might sound more destructive
- Less familiar to general users
- Could imply "patch" rather than reconcile

### Option C: `reconcile`
```bash
plonk reconcile
```

**Pros**:
- Most accurate description
- Clear about comparing and fixing
- Technical users understand it

**Cons**:
- Longer to type
- Less familiar term
- Might sound complex

### Option D: `update`
```bash
plonk update
```

**Pros**:
- Simple, familiar term
- Implies making current
- Short to type

**Cons**:
- Might imply updating plonk itself
- Doesn't convey reconciliation
- Could mean "upgrade" packages

### Option E: `converge`
```bash
plonk converge
```

**Pros**:
- Chef/configuration management pattern
- Implies reaching desired state
- Technically accurate

**Cons**:
- Less familiar
- Might sound mathematical
- Domain-specific term

### Option F: Make it Default
```bash
plonk              # no command = sync/apply
plonk --dry-run    # preview changes
```

**Pros**:
- Shortest possible
- Main operation is default
- Like `make` with no target

**Cons**:
- Breaking change
- Less discoverable
- Conflicts with current status behavior

## Patterns in Other Tools

**Kubernetes**: `kubectl apply -f config.yaml`
- Clear source of truth (file)
- Declarative model

**Ansible**: `ansible-playbook playbook.yml`
- No specific verb, runs the playbook

**Terraform**: `terraform apply`
- Plan then apply model
- Very similar use case

**Puppet**: `puppet apply manifest.pp`
- Apply is standard in config management

**Chef**: `chef-client` or `chef converge`
- Converge to desired state

**Docker Compose**: `docker-compose up`
- Different model (services)

## Related Considerations

1. **Dry Run**: Should we have `--dry-run` or separate `plan` command?
2. **Partial Sync**: `plonk sync --only packages`
3. **Force**: `plonk sync --force` to ignore conflicts?
4. **Auto-sync**: Future feature for watching config?

## Questions for Discussion

1. What mental model do we want users to have?
2. Is consistency with Kubernetes/Terraform valuable?
3. How important is the bidirectional implication of "sync"?
4. Should the command name imply safety or power?

## Recommendation Placeholder

_To be filled after discussion_
