package main

import (
	"os"
	"primjeri/gRPC-LANchat/cli"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	cli.Run()
}
