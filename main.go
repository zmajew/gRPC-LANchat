package main

import (
	node "primjeri/gRPC-LANchat/internal"
)

func main() {
	var n node.Node

	// This should be on Init function
	n.GetOwnLanIp()

	n.Start()

}
