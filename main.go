// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of aforesaid license can be found in the LICENSE file.

package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/teris-io/cli"

	"github.com/choreia/cmd"
)

// This function is executed at the first runtime usage of an imported
// module or (like in this case) before the main() itself is executed.
func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "15:04:05"})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}

func main() {
	// Builds the CLI app with the respective subcommands
	app := cli.New(cmd.Usage).WithCommand(cmd.ExtractMeta)
	// Dispatch the arguments and executes the respective action
	os.Exit(app.Run(os.Args, os.Stdout))
}
