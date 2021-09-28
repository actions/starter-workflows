package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/urfave/cli.v1"
)

var (
	// VersionString is the git describe version set at build time
	VersionString = "?"
	// RevisionString is the git revision set at build time
	RevisionString = "?"
	// RevisionURLString is the full URL to the revision set at build time
	RevisionURLString = "?"
	// GeneratedString is the build date set at build time
	GeneratedString = "?"
	// CopyrightString is the copyright set at build time
	CopyrightString = "?"
)

func init() {
	cli.VersionPrinter = customVersionPrinter
	_ = os.Setenv("VERSION", VersionString)
	_ = os.Setenv("REVISION", RevisionString)
	_ = os.Setenv("GENERATED", GeneratedString)
}

func customVersionPrinter(c *cli.Context) {
	fmt.Printf("%v v=%v rev=%v d=%v go=%v\n", filepath.Base(c.App.Name),
		VersionString, RevisionString, GeneratedString, runtime.Version())
}
