# Template Dotfiles Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Re-introduce env-var-only template dotfile support to the dotfiles module.

**Architecture:** Template files (`.tmpl` extension) in `$PLONK_DIR` are rendered by substituting `{{VAR}}` placeholders with environment variables before being deployed, compared, or diffed. A `lookupEnv` function injected into `DotfileManager` enables testability (consistent with existing `FileSystem` injection pattern).

**Tech Stack:** Go stdlib (`regexp`, `os`), no external dependencies

**Design doc:** `docs/plans/2026-02-05-template-dotfiles-design.md`

---

### Task 1: Add renderTemplate function and isTemplate helper

**Files:**
- Modify: `internal/dotfiles/dotfiles.go` (add constant, helpers, renderTemplate)
- Modify: `internal/dotfiles/dotfiles_test.go` (add tests)

**Step 1: Write the failing tests**

In `internal/dotfiles/dotfiles_test.go`, add at the end of the file:

```go
func TestRenderTemplate(t *testing.T) {
	lookup := func(key string) (string, bool) {
		vars := map[string]string{
			"EMAIL":         "user@example.com",
			"GIT_USER_NAME": "Test User",
		}
		v, ok := vars[key]
		return v, ok
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "single variable",
			input: "email = {{EMAIL}}",
			want:  "email = user@example.com",
		},
		{
			name:  "multiple variables",
			input: "email = {{EMAIL}}\nname = {{GIT_USER_NAME}}",
			want:  "email = user@example.com\nname = Test User",
		},
		{
			name:  "no placeholders",
			input: "just plain text",
			want:  "just plain text",
		},
		{
			name:    "missing variable",
			input:   "email = {{MISSING_VAR}}",
			wantErr: true,
		},
		{
			name:    "multiple missing variables",
			input:   "{{MISSING_ONE}} and {{MISSING_TWO}}",
			wantErr: true,
		},
		{
			name:  "same variable used twice",
			input: "{{EMAIL}} and {{EMAIL}}",
			want:  "user@example.com and user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate([]byte(tt.input), lookup)
			if tt.wantErr {
				if err == nil {
					t.Fatal("renderTemplate() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("renderTemplate() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("renderTemplate() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestIsTemplate(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"gitconfig.tmpl", true},
		{"config/git/config.tmpl", true},
		{"zshrc", false},
		{"tmpl", false},
		{"file.tmpl.bak", false},
	}

	for _, tt := range tests {
		if got := isTemplate(tt.name); got != tt.want {
			t.Errorf("isTemplate(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/dotfiles/ -run "TestRenderTemplate|TestIsTemplate" -v`
Expected: FAIL — `renderTemplate` and `isTemplate` undefined

**Step 3: Write minimal implementation**

In `internal/dotfiles/dotfiles.go`, add `"regexp"` to the imports, then add after the `errSkipDir` var declaration:

```go
const templateExtension = ".tmpl"

var templateVarPattern = regexp.MustCompile(`\{\{([A-Za-z_][A-Za-z0-9_]*)\}\}`)

func isTemplate(name string) bool {
	return strings.HasSuffix(name, templateExtension)
}

func renderTemplate(content []byte, lookupEnv func(string) (string, bool)) ([]byte, error) {
	matches := templateVarPattern.FindAllSubmatch(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	var missing []string
	seen := make(map[string]bool)
	for _, match := range matches {
		varName := string(match[1])
		if seen[varName] {
			continue
		}
		seen[varName] = true
		if _, ok := lookupEnv(varName); !ok {
			missing = append(missing, varName)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}

	result := templateVarPattern.ReplaceAllFunc(content, func(match []byte) []byte {
		varName := string(templateVarPattern.FindSubmatch(match)[1])
		val, _ := lookupEnv(varName)
		return []byte(val)
	})

	return result, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/dotfiles/ -run "TestRenderTemplate|TestIsTemplate" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dotfiles/dotfiles.go internal/dotfiles/dotfiles_test.go
git commit -m "feat(dotfiles): add renderTemplate and isTemplate"
```

---

### Task 2: Add lookupEnv field and strip .tmpl in toTarget

**Files:**
- Modify: `internal/dotfiles/dotfiles.go:22-42` (struct + constructors)
- Modify: `internal/dotfiles/dotfiles.go:346-352` (toTarget method)
- Modify: `internal/dotfiles/dotfiles_test.go` (extend existing test)

**Step 1: Write the failing test**

In `internal/dotfiles/dotfiles_test.go`, add these cases to the existing `TestDotfileManager_ToTarget` test's `tests` slice (after the `bashrc` entry):

```go
{"gitconfig.tmpl", "/home/user/.gitconfig"},
{"config/git/config.tmpl", "/home/user/.config/git/config"},
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/dotfiles/ -run TestDotfileManager_ToTarget -v`
Expected: FAIL — `toTarget("gitconfig.tmpl")` returns `"/home/user/.gitconfig.tmpl"`

**Step 3: Write minimal implementation**

In `internal/dotfiles/dotfiles.go`:

Add `lookupEnv` field to the `DotfileManager` struct (line 22-27):

```go
type DotfileManager struct {
	configDir string
	homeDir   string
	fs        FileSystem
	matcher   *ignore.Matcher
	lookupEnv func(string) (string, bool)
}
```

Set default in `NewDotfileManagerWithFS` (line 35-42):

```go
func NewDotfileManagerWithFS(configDir, homeDir string, ignorePatterns []string, fs FileSystem) *DotfileManager {
	return &DotfileManager{
		configDir: configDir,
		homeDir:   homeDir,
		fs:        fs,
		matcher:   ignore.NewMatcher(ignorePatterns),
		lookupEnv: os.LookupEnv,
	}
}
```

Modify `toTarget` (line 346-352) to strip `.tmpl`:

```go
func (m *DotfileManager) toTarget(relPath string) string {
	// Strip .tmpl extension for template files
	if isTemplate(relPath) {
		relPath = strings.TrimSuffix(relPath, templateExtension)
	}

	// Add dot prefix to the first path component
	parts := strings.SplitN(relPath, string(os.PathSeparator), 2)
	parts[0] = "." + parts[0]
	dotPath := strings.Join(parts, string(os.PathSeparator))
	return filepath.Join(m.homeDir, dotPath)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/dotfiles/ -v`
Expected: ALL PASS (old and new)

**Step 5: Commit**

```bash
git add internal/dotfiles/dotfiles.go internal/dotfiles/dotfiles_test.go
git commit -m "feat(dotfiles): strip .tmpl in toTarget and add lookupEnv field"
```

---

### Task 3: Template-aware Deploy

**Files:**
- Modify: `internal/dotfiles/dotfiles.go:225-268` (Deploy method)
- Modify: `internal/dotfiles/dotfiles_test.go` (add tests)

**Step 1: Write the failing tests**

In `internal/dotfiles/dotfiles_test.go`, add:

```go
func TestDotfileManager_Deploy_Template(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("[user]\n    email = {{EMAIL}}")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		if key == "EMAIL" {
			return "user@example.com", true
		}
		return "", false
	}

	err := m.Deploy("gitconfig.tmpl")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	// Verify rendered content was deployed (not raw template)
	content, ok := fs.Files["/home/user/.gitconfig"]
	if !ok {
		t.Fatal("Deploy() did not create /home/user/.gitconfig")
	}
	want := "[user]\n    email = user@example.com"
	if string(content) != want {
		t.Errorf("Deploy() content = %q, want %q", string(content), want)
	}
}

func TestDotfileManager_Deploy_Template_MissingVar(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("email = {{MISSING_VAR}}")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		return "", false
	}

	err := m.Deploy("gitconfig.tmpl")
	if err == nil {
		t.Fatal("Deploy() expected error for missing variable, got nil")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/dotfiles/ -run "TestDotfileManager_Deploy_Template" -v`
Expected: FAIL — Deploy writes raw template content, not rendered

**Step 3: Write minimal implementation**

In `internal/dotfiles/dotfiles.go`, modify `Deploy` — add template rendering after reading the source content (after line 240 `"failed to read source"`):

```go
// Render template if needed
if isTemplate(name) {
	content, err = renderTemplate(content, m.lookupEnv)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", name, err)
	}
}
```

Insert this block between the `ReadFile` call and the `MkdirAll` call.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/dotfiles/ -run "TestDotfileManager_Deploy" -v`
Expected: PASS (both old and new tests)

**Step 5: Commit**

```bash
git add internal/dotfiles/dotfiles.go internal/dotfiles/dotfiles_test.go
git commit -m "feat(dotfiles): template-aware Deploy"
```

---

### Task 4: Template-aware IsDrifted

**Files:**
- Modify: `internal/dotfiles/dotfiles.go:271-286` (IsDrifted method)
- Modify: `internal/dotfiles/dotfiles_test.go` (add test)

**Step 1: Write the failing test**

In `internal/dotfiles/dotfiles_test.go`, add:

```go
func TestDotfileManager_IsDrifted_Template(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("email = {{EMAIL}}")
	fs.Files["/home/user/.gitconfig"] = []byte("email = user@example.com")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		if key == "EMAIL" {
			return "user@example.com", true
		}
		return "", false
	}

	d := Dotfile{
		Name:   "gitconfig.tmpl",
		Source: "/config/gitconfig.tmpl",
		Target: "/home/user/.gitconfig",
	}

	// Rendered content matches target — not drifted
	drifted, err := m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}
	if drifted {
		t.Error("IsDrifted() = true, want false (rendered content matches)")
	}

	// Change the target — should be drifted
	fs.Files["/home/user/.gitconfig"] = []byte("email = different@example.com")
	drifted, err = m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}
	if !drifted {
		t.Error("IsDrifted() = false, want true (rendered content differs)")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/dotfiles/ -run TestDotfileManager_IsDrifted_Template -v`
Expected: FAIL — IsDrifted compares raw template bytes, not rendered

**Step 3: Write minimal implementation**

In `internal/dotfiles/dotfiles.go`, modify `IsDrifted` — add template rendering after reading source content (after the `ReadFile` call for source, before the `ReadFile` call for target):

```go
// Render template if needed
if isTemplate(d.Name) {
	sourceContent, err = renderTemplate(sourceContent, m.lookupEnv)
	if err != nil {
		return false, fmt.Errorf("failed to render template %s: %w", d.Name, err)
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/dotfiles/ -run "TestDotfileManager_IsDrifted" -v`
Expected: PASS (both old and new tests)

**Step 5: Commit**

```bash
git add internal/dotfiles/dotfiles.go internal/dotfiles/dotfiles_test.go
git commit -m "feat(dotfiles): template-aware IsDrifted"
```

---

### Task 5: Template-aware Diff

**Files:**
- Modify: `internal/dotfiles/dotfiles.go:289-341` (Diff method)
- Modify: `internal/dotfiles/dotfiles_test.go` (add test)

**Step 1: Write the failing test**

In `internal/dotfiles/dotfiles_test.go`, add:

```go
func TestDotfileManager_Diff_Template(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("email = {{EMAIL}}")
	fs.Files["/home/user/.gitconfig"] = []byte("email = old@example.com")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		if key == "EMAIL" {
			return "new@example.com", true
		}
		return "", false
	}

	d := Dotfile{
		Name:   "gitconfig.tmpl",
		Source: "/config/gitconfig.tmpl",
		Target: "/home/user/.gitconfig",
	}

	diff, err := m.Diff(d)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// Diff should show rendered source vs target, not raw template
	if strings.Contains(diff, "{{EMAIL}}") {
		t.Error("Diff() contains raw template placeholder, should contain rendered value")
	}
	if !strings.Contains(diff, "new@example.com") {
		t.Error("Diff() should contain rendered value 'new@example.com'")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/dotfiles/ -run TestDotfileManager_Diff_Template -v`
Expected: FAIL — Diff shows raw template content

**Step 3: Write minimal implementation**

In `internal/dotfiles/dotfiles.go`, modify `Diff` — add template rendering after reading source content (after the `ReadFile` call for source, before the `ReadFile` call for target):

```go
// Render template if needed
if isTemplate(d.Name) {
	sourceContent, err = renderTemplate(sourceContent, m.lookupEnv)
	if err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", d.Name, err)
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/dotfiles/ -run "TestDotfileManager_Diff" -v`
Expected: PASS (both old and new tests)

**Step 5: Commit**

```bash
git add internal/dotfiles/dotfiles.go internal/dotfiles/dotfiles_test.go
git commit -m "feat(dotfiles): template-aware Diff"
```

---

### Task 6: Conflict detection in List

> **Note:** The design doc placed this in `reconcile.go`, but `List()` is the better location — it catches conflicts for all consumers of `List()`, not just `Reconcile()`.

**Files:**
- Modify: `internal/dotfiles/dotfiles.go:45-80` (List method)
- Modify: `internal/dotfiles/dotfiles_test.go` (add test)

**Step 1: Write the failing test**

In `internal/dotfiles/dotfiles_test.go`, add:

```go
func TestDotfileManager_List_TemplateConflict(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Files["/config/gitconfig"] = []byte("plain content")
	fs.Files["/config/gitconfig.tmpl"] = []byte("template content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	_, err := m.List()
	if err == nil {
		t.Fatal("List() expected error for conflicting plain/template files, got nil")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("List() error = %q, want error containing 'conflict'", err.Error())
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/dotfiles/ -run TestDotfileManager_List_TemplateConflict -v`
Expected: FAIL — List returns both files without error

**Step 3: Write minimal implementation**

In `internal/dotfiles/dotfiles.go`, modify `List()`. Replace the final error handling and return (currently lines 74-79):

```go
	// from:
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}

	return dotfiles, err

	// to:
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Check for template/plain file conflicts (same target path)
	targets := make(map[string]string) // target -> source name
	for _, d := range dotfiles {
		if existing, ok := targets[d.Target]; ok {
			return nil, fmt.Errorf("conflict: %s and %s both target %s", existing, d.Name, d.Target)
		}
		targets[d.Target] = d.Name
	}

	return dotfiles, nil
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/dotfiles/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/dotfiles/dotfiles.go internal/dotfiles/dotfiles_test.go
git commit -m "feat(dotfiles): detect template/plain file conflicts in List"
```
