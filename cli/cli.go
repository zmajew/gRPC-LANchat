package cli

import (
	"flag"
	"fmt"
	"primjeri/gRPC-LANchat/internal"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(` insert -volume VOLUME - volume of message notificatin tone`)
}

func (cli *CommandLine) Run() {
	var node internal.Node
	node.Volume = flag.Int("port", 5, "volume of the message tone")

	node.GetOwnLanIp()
	node.Start()
}
