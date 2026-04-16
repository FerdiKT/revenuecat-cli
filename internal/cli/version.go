package cli

import (
	"fmt"
	"os"

	"github.com/FerdiKT/revenuecat-cli/internal/buildinfo"
	"github.com/spf13/cobra"
)

func addVersionCommand(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show build version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(os.Stdout, "revenuecat %s\ncommit=%s\ndate=%s\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
			return err
		},
	})
}
