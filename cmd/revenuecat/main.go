package main

import (
	"os"

	"github.com/FerdiKT/revenuecat-cli/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
