package internal

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"sync"

	ch "primjeri/gRPC-LANchat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Peer struct {
	HostName string
	IP       string
	Address  string
	Client   ch.ChatServiceClient
}

type Node struct {
	IP       string
	HostName string
	Port     string
	Address  string
	PeerBook map[string]*Peer
	mtx      sync.RWMutex
}

func (node *Node) StartListening() {
	listen, err := net.Listen("tcp", node.Address)
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

func (node *Node) AddClient(addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Unable to connect to %s: %v", addr, err)
		return err
	}

	node.mtx.Lock()
	defer node.mtx.Unlock()

	client := ch.NewChatServiceClient(conn)

	node.PeerBook[addr] = new(Peer)
	node.PeerBook[addr].Client = client

	return nil
}

func (node *Node) SetupClient(addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Unable to connect to %s: %v", addr, err)
		return err
	}

	node.mtx.Lock()
	defer node.mtx.Unlock()

	client := ch.NewChatServiceClient(conn)

	//deadline := time.Now().Add(1000 * time.Microsecond)
	//ctx, _ := context.WithDeadline(context.Background(), deadline)

	res, err := client.HandShake(context.Background(), &ch.HandShakeRequest{Name: node.HostName, Address: node.Address})
	if err != nil {
		return err
	}
	node.PeerBook[addr] = new(Peer)
	node.PeerBook[addr].Client = client
	node.PeerBook[addr].HostName = res.Name
	node.PeerBook[addr].IP = res.Ip

	return nil
}

func (node *Node) Start() error {
	node.PeerBook = make(map[string]*Peer)
	node.Port = "4040"
	node.Address = node.IP + ":" + node.Port

	go node.StartListening()

	//var addr string
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

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		m := scanner.Text()
		message := fmt.Sprintf("%s: %s", node.HostName, m)
		for _, peer := range node.PeerBook {
			_, _ = peer.Client.SendMessage(context.Background(), &ch.SendMessageRequest{Mess: message})
			// if err != nil {
			// 	log.Printf("%s did not received message: %v", peer.HostName, err)
			// }
		}

	}
	//fmt.Scanln(&addr)
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
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.)\d{1,3}`)
	match := re.FindAllStringSubmatch(node.IP, -1)
	fmt.Println("Connecting to the chat nodes, please wait...")

	var wg sync.WaitGroup

	for i := 2; i < 256; i++ {
		wg.Add(1)
		ip := match[0][1] + strconv.Itoa(i)

		if ip == node.IP {
			wg.Done()
			continue
		}

		address := ip + ":4040"

		go func() {
			err := node.SetupClient(address)
			if err == nil {
				os.Stderr.WriteString(node.PeerBook[address].HostName + " is online\n")
			}
			wg.Done()
			//fmt.Println(address)
		}()
	}
	wg.Wait()
	fmt.Printf("computers on the chat: %d\n", len(node.PeerBook))
}
