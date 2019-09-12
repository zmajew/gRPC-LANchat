package internal

import (
	"testing"
	"context"
	ch "primjeri/gRPC-LANchat/proto"
)

func TestNodegRPC(t *testing.T) {
node := Node{}

t.Run("SendMessage Test", func(t *testing.T) {
	tests := struct{
		message string
		want bool
	}{
			message: "Hello",
			want: true,
		}

	req := &ch.SendMessageRequest{Mess: tests.message}
	resp, err := node.SendMessage(context.Background(), req)
	if err != nil {
		t.Errorf("TestNodegRPC (SendMessage Test) got unexpected error %v", err)
	}
	if !resp.Received{
		t.Errorf("SendMessage Test (%v)=%v, wanted %v", tests.message, resp.Received, tests.want)
	}
 })

 t.Run("HandShake Test", func(t *testing.T) {
	// tests := struct{
	// 	ip string
	// 	name string
	// 	address string
	// 	wantIP string
	// 	wantName string
	// 	wantAddress string
	// }{
	// 	ip: "192.168.1.2",
	// 	name: "Gaus",
	// 	address: "192.168.1.2:8080",
	// 		wantIP: "192.168.1.2",
	// 		wantName: "Gaus",
	// 		wantAddress: "192.168.1.2:8080",
	// 	}
   node.IP = "192.168.1.2"
   node.HostName = "Gaus"
   node.Address = "192.168.1.2:8080"
   node.PeerBook = make(map[string]*Peer)

//   req := &ch.HandShakeRequest{Ip: tests.ip, Name: tests.name, Address: tests.address,}
   req := &ch.HandShakeRequest{}
	resp, err := node.HandShake(context.Background(), req)
	if err != nil {
		t.Errorf("TestNodegRPC (HandShake Test) got unexpected error %v", err)
	}
	if resp.Wake && (resp.Ip == node.IP) && (resp.Name == node.HostName) && (resp.Address == node.Address) {
		t.Errorf(`HandShake Test: response.Wake = %t wanted: %t\n
									response.IP = %s wanted: %s\n
									response.Name = %s wanted: %s\n
									response.Address = %s wanted: %s\n`, 
									resp.Wake, true, 
									resp.Ip, node.IP,
									resp.Name, node.HostName,
									resp.Address, node.Address,
								 )
	}
 })
}