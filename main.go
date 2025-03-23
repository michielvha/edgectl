/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import "github.com/michielvha/edgectl/cmd"

// Version is set dynamically during build time (pipeline execution works with gitVersion)
var Version = "dev"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
