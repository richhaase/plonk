package commands

// MockGit implements GitInterface for testing
type MockGit struct {
	CloneFunc  func(repoURL, targetDir string) error
	PullFunc   func(repoDir string) error
	IsRepoFunc func(dir string) bool
}

func (m *MockGit) Clone(repoURL, targetDir string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(repoURL, targetDir)
	}
	return nil
}

func (m *MockGit) Pull(repoDir string) error {
	if m.PullFunc != nil {
		return m.PullFunc(repoDir)
	}
	return nil
}

func (m *MockGit) IsRepo(dir string) bool {
	if m.IsRepoFunc != nil {
		return m.IsRepoFunc(dir)
	}
	return false
}