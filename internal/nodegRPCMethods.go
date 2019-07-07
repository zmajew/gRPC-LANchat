package internal

import (
	"context"
	"fmt"
	"os"
	ch "primjeri/gRPC-LANchat/proto"
)

func (node *Node) SendMessage(ctx context.Context, stream *ch.SendMessageRequest) (*ch.SendMessageResponse, error) {
	fmt.Println(stream.Mess)
	return &ch.SendMessageResponse{Received: true}, nil
}

func (node *Node) HandShake(ctx context.Context, knocknock *ch.HandShakeRequest) (*ch.HandShakeResponse, error) {
	os.Stderr.WriteString(knocknock.Name + " is online\n")
	return &ch.HandShakeResponse{Ip: node.IP, Name: node.HostName}, nil
}
