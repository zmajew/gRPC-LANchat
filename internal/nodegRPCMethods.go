package internal

import (
	"context"
	"fmt"
	"os"
	ch "primjeri/gRPC-LANchat/proto"
)

func (node *Node) SendMessage(ctx context.Context, stream *ch.SendMessageRequest) (*ch.SendMessageResponse, error) {
	go node.BeepMessage()
	fmt.Println(stream.Mess)
	return &ch.SendMessageResponse{Received: true}, nil
}

func (node *Node) HandShake(ctx context.Context, knocknock *ch.HandShakeRequest) (*ch.HandShakeResponse, error) {
	os.Stderr.WriteString(knocknock.Name + " is online\n")

	address := knocknock.Address

	node.PeerBook[address] = new(Peer)
	node.PeerBook[address].HostName = knocknock.Name

	err := node.AddClient(address)
	if err != nil {
		os.Stderr.WriteString(node.PeerBook[address].HostName + " is not online\n")
	}

	return &ch.HandShakeResponse{Ip: node.IP, Name: node.HostName, Address: node.Address}, nil
}
