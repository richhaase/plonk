# Phase 5: Lock v2 & Hooks Implementation

## Objective
Implement the extensible lock file v2 format and hook system to prepare plonk for AI Lab features while maintaining backward compatibility.

## Timeline
Day 8 (8 hours)

## Current State
- Lock file v1 only supports packages
- No hook system exists
- ~13,826 LOC after Phase 4
- Resource abstraction in place and working

## Target State
- Lock file v2 with generic `resources` section
- Automatic v1→v2 migration on write
- Working hook system with pre/post sync
- Full backward compatibility
- Clear path for future resource types

## Task Breakdown

### Task 5.1: Define Lock v2 Schema (1 hour)
**Agent Instructions:**
1. Create `internal/lock/schema_v2.go`:
   ```go
   package lock

   const CurrentVersion = 2

   type LockV2 struct {
       Version   int                    `yaml:"version"`
       Packages  map[string][]Package   `yaml:"packages,omitempty"`  // For compatibility
       Resources []ResourceEntry        `yaml:"resources,omitempty"` // New generic section
   }

   type ResourceEntry struct {
       Type       string                 `yaml:"type"`       // "package", "dotfile", "docker-compose"
       ID         string                 `yaml:"id"`         // Resource-specific identifier
       State      string                 `yaml:"state"`      // "managed", "missing", etc.
       Metadata   map[string]interface{} `yaml:"metadata"`   // Resource-specific data
       InstalledAt string                `yaml:"installed_at,omitempty"`
   }
   ```

2. Keep existing Package struct for backward compatibility

3. Commit: "feat: define lock file v2 schema with resources section"

**Validation:**
- Schema supports both old packages and new resources
- Clean separation of concerns

### Task 5.2: Implement Lock File Migration (2 hours)
**Agent Instructions:**
1. Update `internal/lock/yaml_lock.go`:
   ```go
   func (l *YAMLLock) Read() (*LockData, error) {
       // Read raw YAML to determine version
       var raw map[string]interface{}
       // ... read logic ...

       version := 1
       if v, ok := raw["version"].(int); ok {
           version = v
       }

       switch version {
       case 1:
           return l.readV1(raw)
       case 2:
           return l.readV2(raw)
       default:
           return nil, fmt.Errorf("unsupported lock version: %d", version)
       }
   }

   func (l *YAMLLock) Write(data *LockData) error {
       // Always write as v2
       v2Data := l.migrateToV2(data)
       // Log migration if it occurred
       if data.Version < CurrentVersion {
           log.Printf("Migrated lock file from v%d to v%d", data.Version, CurrentVersion)
       }
       return l.writeV2(v2Data)
   }
   ```

2. Implement migration logic:
   - Convert v1 packages to v2 resources format
   - Preserve all existing data
   - Set version = 2

3. Add tests for:
   - Reading v1 files
   - Reading v2 files
   - Migration from v1 to v2
   - Round-trip compatibility

4. Commit: "feat: implement lock file v1 to v2 migration"

**Validation:**
- Existing lock files still work
- New writes use v2 format
- No data loss during migration

### Task 5.3: Update Resource Integration (1.5 hours)
**Agent Instructions:**
1. Update orchestrator to write resources to lock file:
   ```go
   func (o *Orchestrator) writeLock(resources []resources.Resource) error {
       lockData := &lock.LockData{
           Version: lock.CurrentVersion,
           Resources: []lock.ResourceEntry{},
       }

       for _, res := range resources {
           items := res.Actual(o.ctx)
           for _, item := range items {
               if item.State == "managed" {
                   entry := lock.ResourceEntry{
                       Type:  res.ID(),
                       ID:    item.Name,
                       State: item.State,
                       Metadata: map[string]interface{}{
                           // Resource-specific metadata
                       },
                       InstalledAt: time.Now().Format(time.RFC3339),
                   }
                   lockData.Resources = append(lockData.Resources, entry)
               }
           }
       }

       return o.lock.Write(lockData)
   }
   ```

2. Update lock reading to populate resources

3. Ensure backward compatibility:
   - Package managers still work with both formats
   - Dotfiles are tracked in resources section

4. Commit: "feat: integrate resources with lock v2"

**Validation:**
- `plonk sync` creates v2 lock files
- Resources are properly tracked
- No regression in functionality

### Task 5.4: Implement Hook System (2.5 hours)
**Agent Instructions:**
1. Create `internal/orchestrator/hooks.go`:
   ```go
   package orchestrator

   import (
       "context"
       "os/exec"
       "time"
   )

   type HookRunner struct {
       defaultTimeout time.Duration
   }

   func NewHookRunner() *HookRunner {
       return &HookRunner{
           defaultTimeout: 10 * time.Minute,
       }
   }

   func (h *HookRunner) Run(ctx context.Context, hooks []config.Hook, phase string) error {
       for _, hook := range hooks {
           if hook.Phase != phase {
               continue
           }

           timeout := h.defaultTimeout
           if hook.Timeout != "" {
               if d, err := time.ParseDuration(hook.Timeout); err == nil {
                   timeout = d
               }
           }

           ctx, cancel := context.WithTimeout(ctx, timeout)
           defer cancel()

           cmd := exec.CommandContext(ctx, "sh", "-c", hook.Command)
           output, err := cmd.CombinedOutput()

           if err != nil {
               if !hook.ContinueOnError {
                   return fmt.Errorf("hook failed: %s\n%s", err, output)
               }
               log.Printf("Hook failed (continuing): %s\n%s", err, output)
           }
       }
       return nil
   }
   ```

2. Update config types in `internal/config/types.go`:
   ```go
   type Config struct {
       // ... existing fields ...
       Hooks Hooks `yaml:"hooks,omitempty"`
   }

   type Hooks struct {
       PreSync  []Hook `yaml:"pre_sync,omitempty"`
       PostSync []Hook `yaml:"post_sync,omitempty"`
   }

   type Hook struct {
       Command         string `yaml:"command"`
       Timeout         string `yaml:"timeout,omitempty"`         // e.g., "30s", "5m"
       ContinueOnError bool   `yaml:"continue_on_error,omitempty"`
       Phase           string // Set by unmarshaling location
   }
   ```

3. Integrate hooks into orchestrator sync:
   ```go
   func (o *Orchestrator) Sync() error {
       // Run pre-sync hooks
       if err := o.hookRunner.Run(o.ctx, o.config.Hooks.PreSync, "pre_sync"); err != nil {
           return fmt.Errorf("pre-sync hook failed: %w", err)
       }

       // ... existing sync logic ...

       // Run post-sync hooks
       if err := o.hookRunner.Run(o.ctx, o.config.Hooks.PostSync, "post_sync"); err != nil {
           return fmt.Errorf("post-sync hook failed: %w", err)
       }

       return nil
   }
   ```

4. Commit: "feat: implement hook system with pre/post sync support"

**Validation:**
- Hooks execute at correct times
- Timeout works correctly
- continue_on_error behavior works

### Task 5.5: Add Hook Tests and Documentation (1 hour)
**Agent Instructions:**
1. Add hook tests:
   - Test successful hook execution
   - Test hook timeout
   - Test continue_on_error
   - Test invalid commands

2. Update example config with hooks:
   ```yaml
   # plonk.yaml
   packages:
     homebrew:
       - jq

   hooks:
     pre_sync:
       - command: "echo 'Starting plonk sync...'"
       - command: "./scripts/backup.sh"
         timeout: "2m"
     post_sync:
       - command: "./scripts/notify.sh"
         continue_on_error: true
   ```

3. Add hook documentation comments

4. Commit: "test: add hook system tests and examples"

**Validation:**
- All hook tests pass
- Example demonstrates usage

### Task 5.6: Final Validation (1 hour)
**Agent Instructions:**
1. Test complete flow:
   ```bash
   # Create test config with hooks
   cat > test-plonk.yaml << EOF
   packages:
     homebrew:
       - jq
   hooks:
     pre_sync:
       - command: "echo 'Pre-sync hook running'"
     post_sync:
       - command: "echo 'Post-sync hook complete'"
   EOF

   # Run sync
   plonk sync -c test-plonk.yaml

   # Verify lock file is v2
   cat plonk.lock | grep "version: 2"
   ```

2. Test v1→v2 migration:
   - Create old v1 lock file
   - Run sync
   - Verify migration message in output
   - Verify v2 format

3. Run all tests:
   ```bash
   go test ./...
   just test-ux
   ```

4. Create summary report

5. Commit: "chore: validate Phase 5 implementation"

## Risk Mitigations

1. **Breaking Lock Compatibility**: Reader supports both v1 and v2
2. **Hook Security**: Only run user-configured commands
3. **Hook Failures**: Default fail-fast, optional continue
4. **Migration Issues**: Log all migrations, preserve all data

## Success Criteria
- [ ] Lock v2 schema implemented and tested
- [ ] Automatic v1→v2 migration works
- [ ] Hook system executes pre/post sync
- [ ] All existing functionality preserved
- [ ] Tests pass with no regressions
- [ ] Clear foundation for future resources

## Notes for Agents
- Keep changes focused on lock v2 and hooks
- Don't modify Resource interface
- Maintain backward compatibility
- Test migration thoroughly
- Keep hook system simple and secure
