package main

import (
	"os"

	gitrestoremtime "github.com/gartnera/git-restore-mtime"
	"github.com/spf13/cobra"
)

const (
	maxDepthArg = "max-depth"
)

func init() {
	flags := rootCmd.Flags()
	flags.Int(maxDepthArg, 0, "maximum depth to traverse the commit history (default unlimited)")
}

var rootCmd = &cobra.Command{
	Use:          "git-restore-mtime <path>",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := []gitrestoremtime.ManagerOptionT{}
		if maxDepth, _ := cmd.Flags().GetInt(maxDepthArg); maxDepth > 0 {
			opts = append(opts, gitrestoremtime.WithMaxDepth(maxDepth))
		}
		m, err := gitrestoremtime.NewManagerFromPath(args[0], opts...)
		if err != nil {
			return err
		}
		return m.RunDefault()
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
