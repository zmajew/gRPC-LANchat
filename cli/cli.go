package cli

import (
	"flag"
	"fmt"
	"primjeri/gRPC-LANchat/internal"
	"runtime"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(` insert -volume VOLUME - volume of the message notificatin tone ( 0 < VOLUME < 100)`)
}

func (cli *CommandLine) Run() {
	var node internal.Node
	volume := flag.Int("volume", 5, "volume of the message notificatin tone ( 0 < VOLUME < 100)")
	flag.Parse()
	if *volume < 0 || *volume > 100 {
		flag.Usage()
		runtime.Goexit()
		//*volume = 5
	}
	node.Volume = volume

	node.GetOwnLanIp()
	node.Start()
}
