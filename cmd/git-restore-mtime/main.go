package main

import (
	"os"
	"runtime/pprof"

	gitrestoremtime "github.com/gartnera/git-restore-mtime"
	"github.com/spf13/cobra"
)

const (
	maxDepthArg   = "max-depth"
	cpuProfileArg = "cpu-profile"
	memProfileArg = "mem-profile"
)

func init() {
	flags := rootCmd.Flags()
	flags.Int(maxDepthArg, 0, "maximum depth to traverse the commit history (default unlimited)")

	persistentFlags := rootCmd.PersistentFlags()
	persistentFlags.String(cpuProfileArg, "", "path to write pprof cpu profile")
	persistentFlags.String(memProfileArg, "", "path to write pprof memory profile")
}

var rootCmd = &cobra.Command{
	Use:          "git-restore-mtime <path>",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cpuProfile, _ := cmd.Flags().GetString(cpuProfileArg)
		if cpuProfile != "" {
			f, err := os.Create(cpuProfile)
			if err != nil {
				return err
			}
			return pprof.StartCPUProfile(f)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		pprof.StopCPUProfile()
		memProfile, _ := cmd.Flags().GetString(memProfileArg)
		if memProfile != "" {
			f, err := os.Create(memProfile)
			if err != nil {
				return err
			}
			defer f.Close()
			return pprof.WriteHeapProfile(f)
		}
		return nil
	},
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
