package internal

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"

	ch "primjeri/gRPC-LANchat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Node struct {
	Name    string
	IP string
	Port string
	Peer    map[string]ch.ChatServiceClient
}

func (node *Node) SendMessage(ctx context.Context, stream *ch.Request) (*ch.Response, error) {
	fmt.Println(stream.Mess)
	return &ch.Response{Received: true}, nil
}

func (node *Node) StartListening() {
	address := node.IP + ":" + node.Port
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	//grpcServer := grpc.NewServer()

	ch.RegisterChatServiceServer(grpcServer, node)
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (node *Node) SetupClient(addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())

	if err != nil {
		log.Printf("Unable to connect to %s: %v", addr, err)
		return err
	}

	node.Peer[addr] = ch.NewChatServiceClient(conn)

	_, err = node.Peer[addr].SendMessage(context.Background(), &ch.Request{Mess: node.Name})
	if err != nil {
		log.Printf("Error making request to %s: %v", addr, err)
		return err
	}

	return nil
}

func (node *Node) Start() error {
	node.Peer = make(map[string]ch.ChatServiceClient)

	address := node.IP + ":" + node.Port
	fmt.Println("Your chat addres is:", address)

	go node.StartListening()

	var addr string
	var again string

	fmt.Printf("Enter the address to chat with (example: 192.168.1.2:4040): ")

	Loop1:
	for {
		fmt.Scanln(&addr)
		if err := node.SetupClient(addr); err != nil {
			fmt.Printf("Unable to setup connection with %s: %v\n", addr, err)
			
			for {
			fmt.Printf("Do you want to try again [y/n]: ")
			fmt.Scanln(&again)
			switch again {
			case "y":
				fmt.Printf("Enter the address to chat with (example: 192.168.1.2:4040): ")
				continue Loop1
			case "n":
				return nil
			default:
				continue
			}
		}
		} else {
			break
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		m := scanner.Text()
		message := fmt.Sprintf("%s: %s", node.Name, m)
		_, err := node.Peer[addr].SendMessage(context.Background(), &ch.Request{Mess: message})
		if err != nil {
			log.Printf("%s did not received message: %v", node.Peer, err)
		}
	}
	return nil
}

func (node *Node) GetOwnLanIp() error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		//os.Exit(1)
		return err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				//os.Stdout.WriteString(ip + "\n")
				if ip[:8] == "192.168." {
					node.IP = ip
				}		
			}
		}
	}
	return nil
}
