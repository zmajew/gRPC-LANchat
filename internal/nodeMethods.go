package internal

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"

	ch "primjeri/gRPC-LANchat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Peer struct {
	HostName string
	IP       string
	Client   ch.ChatServiceClient
}

type Node struct {
	IP       string
	HostName string
	Port     string
	PeerBook map[string]*Peer
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

	a := ch.NewChatServiceClient(conn)

	node.PeerBook[addr] = new(Peer)
	node.PeerBook[addr].Client = a

	res, err := node.PeerBook[addr].Client.HandShake(context.Background(), &ch.HandShakeRequest{})
	if err != nil {
		return err
	}
	node.PeerBook[addr].HostName = res.Name
	node.PeerBook[addr].IP = res.Ip

	return nil
}

func (node *Node) Start() error {
	node.PeerBook = make(map[string]*Peer)
	node.Port = "4040"

	address := node.IP + ":" + node.Port
	fmt.Println("Your chat addres is:", address)

	go node.StartListening()

	var addr string
	node.ScanLan()
	// 	var again string

	// 	fmt.Printf("Enter the address to chat with (example: 192.168.1.2:4040): ")

	// Loop1:
	// 	for {
	// 		fmt.Scanln(&addr)
	// 		if err := node.SetupClient(addr); err != nil {
	// 			fmt.Printf("Unable to setup connection with %s: %v\n", addr, err)

	// 			for {
	// 				fmt.Printf("Do you want to try again [y/n]: ")
	// 				fmt.Scanln(&again)
	// 				switch again {
	// 				case "y":
	// 					fmt.Printf("Enter the address to chat with (example: 192.168.1.2:4040): ")
	// 					continue Loop1
	// 				case "n":
	// 					return nil
	// 				default:
	// 					continue
	// 				}
	// 			}
	// 		} else {
	// 			break
	// 		}
	// 	}

	// 	scanner := bufio.NewScanner(os.Stdin)
	// 	for scanner.Scan() {
	// 		m := scanner.Text()
	// 		message := fmt.Sprintf("%s: %s", node.Name, m)
	// 		_, err := node.PeerBook[addr].Client.SendMessage(context.Background(), &ch.SendMessageRequest{Mess: message})
	// 		if err != nil {
	// 			log.Printf("%s did not received message: %v", node.PeerBook[addr].HostName, err)
	// 		}
	// 	}
	fmt.Scanln(&addr)
	return nil
}

func (node *Node) GetOwnLanIp() error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		//os.Exit(1)
		return err
	}

	hostaName, err := os.Hostname()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		//os.Exit(1)
		return err
	}
	node.HostName = hostaName

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

func (node *Node) ScanLan() {
	for i := 2; i < 6; i++ {
		re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.)\d{1,3}`)
		match := re.FindAllStringSubmatch(node.IP, -1)
		address := match[0][1] + strconv.Itoa(i) + ":4040"

		go func() {
			err := node.SetupClient(address)
			if err == nil {
				fmt.Println(node.PeerBook[address].HostName)
			}
		}()
	}

}
