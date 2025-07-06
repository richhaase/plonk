package commands

import (
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Generate plonk.yaml from existing shell environment",
	Long: `Import existing shell environment configuration to create plonk.yaml.
Discovers installed packages from:
- Homebrew (brew list)
- ASDF (asdf list)  
- NPM (npm list -g)

Copies dotfiles:
- .zshrc, .gitconfig, .zshenv

Generates a complete plonk.yaml configuration file.`,
	RunE: runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	return nil
}
