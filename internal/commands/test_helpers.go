package commands

// MockGit implements GitInterface for testing
type MockGit struct {
	CloneFunc  func(repoURL, targetDir string) error
	PullFunc   func(repoDir string) error
	IsRepoFunc func(dir string) bool
}

// MockGitWithBranch implements GitInterface with branch support for testing
type MockGitWithBranch struct {
	CloneBranchFunc func(repoURL, targetDir, branch string) error
	PullFunc        func(repoDir string) error
	IsRepoFunc      func(dir string) bool
}

func (m *MockGitWithBranch) Clone(repoURL, targetDir string) error {
	return m.CloneBranch(repoURL, targetDir, "")
}

func (m *MockGitWithBranch) CloneBranch(repoURL, targetDir, branch string) error {
	if m.CloneBranchFunc != nil {
		return m.CloneBranchFunc(repoURL, targetDir, branch)
	}
	return nil
}

func (m *MockGitWithBranch) Pull(repoDir string) error {
	if m.PullFunc != nil {
		return m.PullFunc(repoDir)
	}
	return nil
}

func (m *MockGitWithBranch) IsRepo(dir string) bool {
	if m.IsRepoFunc != nil {
		return m.IsRepoFunc(dir)
	}
	return false
}

func (m *MockGit) Clone(repoURL, targetDir string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(repoURL, targetDir)
	}
	return nil
}

func (m *MockGit) CloneBranch(repoURL, targetDir, branch string) error {
	// For basic MockGit, just call Clone
	return m.Clone(repoURL, targetDir)
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