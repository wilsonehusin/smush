package main

import (
	"fmt"
	"runtime/debug"
)

var version = "dev"
var vcs = ""

func printVersion() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Unable to retrieve build information")
	}
	fmt.Printf("%s\n", version)
	for _, setting := range bi.Settings {
		switch setting.Key {
		case "vcs.revision":
			fmt.Printf("%s\n", setting.Value)
		case "vcs.modified":
			if setting.Value == "true" {
				fmt.Printf("PREVIEW BUILD (dirty)\n")
			}
		}
	}
	fmt.Printf("%s\n", bi.GoVersion)
}
