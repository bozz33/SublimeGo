package main

import (
	"os"

	"github.com/bozz33/SublimeGo/cmd/sublimego/commands"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func main() {
	// Inject build info into commands
	commands.SetVersionInfo(Version, GitCommit, BuildDate)

	// Execute root command
	if err := commands.Execute(); err != nil {
		// Afficher l'erreur avant de quitter
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
		os.Exit(1)
	}
}
