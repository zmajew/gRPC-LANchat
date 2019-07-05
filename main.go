package main

import (
	"fmt"
	node "primjeri/gRPC-LANchat/internal"
)

func main() {
	var n node.Node
	fmt.Printf("Enter your node name: ")
	fmt.Scanln(&n.Name)
	fmt.Printf("Enter your node port: ")
	fmt.Scanln(&n.Port)
	n.GetOwnLanIp()

	n.Start()
}
