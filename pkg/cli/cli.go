package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

func WrapRunE(log *logger.Logger, fn func(log *logger.Logger, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := fn(log, cmd, args); err != nil {
			log.Errorf("ERROR: %v", err)
			os.Exit(1)
		}
	}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustGetString(cmd *cobra.Command, name string) string {
	val, err := cmd.Flags().GetString(name)
	Must(err)
	return val
}

func MustGetBool(cmd *cobra.Command, name string) bool {
	val, err := cmd.Flags().GetBool(name)
	Must(err)
	return val
}

func MustGetInt(cmd *cobra.Command, name string) int {
	val, err := cmd.Flags().GetInt(name)
	Must(err)
	return val
}

func MustGetStringArray(cmd *cobra.Command, name string) []string {
	val, err := cmd.Flags().GetStringArray(name)
	Must(err)
	return val
}

func GetBuildInfo() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "(no build info available)"
	}
	commit := "unknown commit"
	commitDate := "unknown date"
	dirty := ""
	for _, setting := range bi.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value[:8]
		case "vcs.time":
			commitDate = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				dirty = " (dirty)"
			}
		}
	}
	return fmt.Sprintf("revision: %s (%s)%s", commit, commitDate, dirty)
}

func GetRelativePath(p string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return p
	}
	relP, err := filepath.Rel(cwd, p)
	if err != nil {
		return p
	}
	return relP
}

func GetAbsolutePath(p string) string {
	absP, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return absP
}
