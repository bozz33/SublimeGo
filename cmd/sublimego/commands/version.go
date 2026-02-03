package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Display version information",
	Long:    `Display the version, git commit, and build date of SublimeGo.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("SublimeGo %s\n", version)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		fmt.Printf("Build Date: %s\n", buildDate)
	},
}
