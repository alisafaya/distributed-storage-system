package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"peer"
	"strconv"
	"strings"
)

func readline() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSuffix(text, "\n")
}

func startServer() {
	// include other socket connections functionalities
	p := new(peer.Peer)
	rpc.Register(p)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+peer.Me.Port)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func handleOptions() {
	for {
		fmt.Print(">Please select an option: ")
		option := readline()

		switch option {
		case "1":
			fmt.Print(">Enter peer address to connect: ")
			address := readline()
			peer.Connect(address)
			fmt.Println(">(Response): Connection established.")
		case "2":
			fmt.Print(">Enter the key to find its successor: ")
			key, _ := strconv.Atoi(readline())
			_peer := peer.GetSuccessor(uint32(key))
			println(">(Response):", _peer.Id, _peer.Ip, _peer.Port)
		case "3":
			fmt.Println(">Enter the filename to take its hash: ")
			filehash := peer.Hash(readline())
			fmt.Println(">(Response):", filehash)
		case "4":
			fmt.Println(">Display my-id, succ-id, and pred-id: ")
			fmt.Println(">(Response):", peer.Me.Id, peer.Succ.Id, peer.Pred.Id)
		case "5":
			fmt.Println(">Display the stored filenames and their keys: ")
			for k, v := range peer.Files {
				fmt.Println("hash:", k, "file name:", v)
			}
		case "6":
			peer.Leave()
			fmt.Println(">Exiting...")
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func main() {
	port := os.Args[1]
	peer_ip := peer.GetOutboundIP()
	peer_id := peer.Hash(peer_ip + ":" + port)

	println("Current peer informantion, (peer_id, peer_ip, port) :", peer_id, peer_ip, port)

	peer.Me = &peer.Peer{Id: peer_id, Ip: peer_ip, Port: port}
	peer.Succ = &peer.Peer{Id: peer_id, Ip: peer_ip, Port: port}
	peer.Pred = &peer.Peer{Id: peer_id, Ip: peer_ip, Port: port}
	// peer.Pred = nil
	startServer()

	fmt.Println(`
	1) Enter the peer address to connect:
	2) Enter the key to find its successor:
	3) Enter the filename to take its hash:
	4) Display my-id, succ-id, and pred-id:
	5) Display the stored filenames and their keys:
	6) Exit:`)

	handleOptions()
}
