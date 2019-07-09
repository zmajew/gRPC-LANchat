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

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				if ip[:8] == "192.168." {
					node.IP = ip
				}
			}
		}
	}
	return nil
}

func (node *Node) scanLan() {
	fmt.Println("Connecting to the chat nodes, please wait...")

	ips := getLanIPs()
	if len(ips) == 0 {
		fmt.Println("LAN is empty, waiting...")
		return
	}

	var wg0 sync.WaitGroup
	go func() {
		wg0.Add(1)
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
				}
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Printf("computers on the chat: %d\n", len(node.PeerBook))
	}()
}

func getLanIPs() []*string {
	os := runtime.GOOS
	var out []byte
	var err error
	rersponse := []*string{}

	switch os {
	case "windows":
		out, err = exec.Command("arp", "-a").Output()
		errorCheck(err)
	case "linux":
		out, err = exec.Command("arp", "-n").Output()
		errorCheck(err)
	case "darwin":
		out, err = exec.Command("arp", "-a").Output()
		errorCheck(err)
	}

	temp := strings.Split(string(out), "\n")
	for _, v := range temp {
		if strings.Contains(v, "dynamic") {
			ip := extractIP(v)
			if isNotRouter(ip) {
				rersponse = append(rersponse, &ip)
			}
		}
	}
	return rersponse
}

func errorCheck(err error) {
	if err != nil {
		fmt.Println("Greska:", err)
	}
}

func extractIP(s string) string {
	r := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
	ip := r.FindString(s)

	return ip
}

func isNotRouter(s string) bool {
	if s[len(s)-2:] == ".1" {
		return false
	}
	return true
}
