package main

import "github.com/rpcarvs/faz/cmd"

// version is injected at build time for tagged releases.
var version string

// main executes the root CLI command tree.
func main() {
	cmd.Execute(version)
}
