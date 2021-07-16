package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/deref/exo/cmdutil"
	"github.com/deref/exo/jsonutil"
	"github.com/deref/exo/osutil"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deamonCmd)
}

var deamonCmd = &cobra.Command{
	Hidden: true,
	Use:    "deamon",
	Short:  "Start the exo deamon",
	Long: `Start the exo deamon and then do nothing else.

Since most commands implicitly start the exo deamon, users generally do not
have to invoke this themselves.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ensureDeamon()
		return nil
	},
}

var runState struct {
	Pid int    `json:"pid"`
	URL string `json:"url"`
}

func ensureDeamon() {
	paths := cmdutil.MustMakeDirectories()

	// Validate exod process record.
	err := loadRunState(paths.RunStateFile)
	running := false
	switch {
	case err == nil:
		running = osutil.IsValidPid(runState.Pid)
		if !running {
			_ = os.Remove(paths.RunStateFile)
		}
	case os.IsNotExist(err):
		// Not running.
	default:
		cmdutil.Fatalf("checking run state: %w", err)
	}

	if running {
		// TODO: health check.
		return
	}

	// Start server in background.
	exoPath := os.Args[0]
	cmd := exec.Command(exoPath, "server")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	if err := cmd.Start(); err != nil {
		cmdutil.Fatalf("starting exo server: %w", err)
	}

	// Write run state.
	runState.Pid = cmd.Process.Pid
	runState.URL = fmt.Sprintf("http://%s/", cmdutil.GetAddr())
	if err := jsonutil.MarshalFile(paths.RunStateFile, runState); err != nil {
		cmdutil.Fatalf("writing run state: %w", err)
	}
}

func loadRunState(path string) error {
	return jsonutil.UnmarshalFile(path, &runState)
}
