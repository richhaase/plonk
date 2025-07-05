package managers

import "bytes"

// CommandRunner provides common command execution functionality
type CommandRunner struct {
	executor    CommandExecutor
	commandName string
}

// NewCommandRunner creates a new command runner for a specific command
func NewCommandRunner(executor CommandExecutor, commandName string) *CommandRunner {
	return &CommandRunner{
		executor:    executor,
		commandName: commandName,
	}
}

// RunCommandWithOutput executes a command and returns output + error
func (c *CommandRunner) RunCommandWithOutput(args ...string) (string, error) {
	cmd := c.executor.Execute(c.commandName, args...)
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	err := cmd.Run()
	return out.String(), err
}

// RunCommand executes a command and returns success/error (ignores output)
func (c *CommandRunner) RunCommand(args ...string) error {
	_, err := c.RunCommandWithOutput(args...)
	return err
}