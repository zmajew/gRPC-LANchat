package internal

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
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
	Volume   *int
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
	node.Port = "4041"
	node.Address = node.IP + ":" + node.Port

	go node.StartListening()

	node.scanLan()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		m := scanner.Text()
		message := fmt.Sprintf("%s: %s", node.HostName, m)
		for _, peer := range node.PeerBook {
			_, _ = peer.Client.SendMessage(context.Background(), &ch.SendMessageRequest{Mess: message})
		}

	}
	return nil
}

func (node *Node) GetOwnLanIp() error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		return err
	}

	hostName, err := os.Hostname()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		return err
	}
	node.HostName = hostName

	localIPs := []string{}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				if ip[:8] == "192.168." && ip[len(ip)-1:] != "1" {
					localIPs = append(localIPs, ip)
				}
			}
		}
	}

	if len(localIPs) > 1 {
		number := 0
		fmt.Println("Your computer has several LAN interfaces: ")
		for i, v := range localIPs {
			fmt.Printf("%d.   %s\n", i+1, v)
		}
		fmt.Println("Enter the ordinal number of your IP from the list:")
		for {
			_, err := fmt.Scan(&number)
			if err != nil || number > len(localIPs) || number < 1 {
				fmt.Printf("Enter number in the range %d - %d:\n", 1, len(localIPs)+1)
				continue
			}
			break
		}
		node.IP = localIPs[number-1]

		return nil
	}

	if len(localIPs) == 0 {
		fmt.Println("Cannot determine host IP. Enter your LAN IP:")
		ip := ""
		for {
			_, err := fmt.Scan(&ip)
			if err != nil {
				continue
			}
			break
		}
		node.IP = ip

		return nil
	}

	node.IP = localIPs[0]
	return nil
}

func (node *Node) scanLan() {
	fmt.Println("Your LAN address is:", node.Address)
	fmt.Println("Connecting to the chat nodes, please wait...")

	ips, err := getLanIPs()
	errorCheck(err)
	if len(ips) == 0 {
		fmt.Println("LAN is empty, waiting...")
		fmt.Println("Check if yours or others firewals blocks the trafic, or check if you are on the same domain with the other computers")
		return
	}

	go func() {
		var wg sync.WaitGroup
		for _, ip := range ips {
			wg.Add(1)

			if *ip == node.IP {
				wg.Done()
				continue
			}

			address := *ip + ":4041"
			go func() {
				err := node.SetupClient(address)
				if err == nil {
					os.Stderr.WriteString(node.PeerBook[address].HostName + " is online\n")
				} // else {
				// 	fmt.Println(err)
				// }
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Printf("computers on the chat: %d\n", len(node.PeerBook))
	}()
}

func getLanIPs() ([]*string, error) {
	os := runtime.GOOS
	var out []byte
	var err error
	response := []*string{}

	switch os {
	case "windows":
		out, err = exec.Command("arp", "-a").Output()
		if err != nil {
			return response, err
		}
		if err != nil {
			fmt.Println(`Chack if the C:\WINDOWS\SYSTEM32 is added to the Enviroment variables`)
		}
	case "linux":
		out, err = exec.Command("arp", "-n").Output()
		if err != nil {
			return response, err
		}
	case "darwin":
		out, err = exec.Command("arp", "-a").Output()
		if err != nil {
			return response, err
		}
	}

	temp := strings.Split(string(out), "\n")
	for _, v := range temp {
		if strings.Contains(v, "dynamic") || strings.Contains(v, "ether") {
			ip := extractIP(v)
			if isNotRouter(ip) {
				response = append(response, &ip)
			}
		}
	}
	return response, nil
}

func errorCheck(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func extractIP(s string) string {
	r := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
	ip := r.FindString(s)

	return ip
}

func isNotRouter(s string) bool {
	if s != "" {
		if s[len(s)-2:] == ".1" || s[len(s)-4:] == ".255" || s[len(s)-2:] == ".0" {
			return false
		}
		return true
	}
	return false
}
