package main

import (
	"bufio"
	"fmt"
	"os"
	"peer"
	"strings"
)

func readline() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSuffix(text, "\n")
}

func serverMessage(m string) {
	fmt.Println(">Server Response:", m)
}

var ip string
var port string

func searchFile(filename string) (*peer.Peer, error) {
	file_id := peer.Hash(filename)
	node := &peer.Peer{}
	conn := peer.GetConnection(ip + ":" + port)
	err := peer.GetSuccessorOf(conn, file_id, node)
	return node, err
}

func store(filename string) {
	node, _ := searchFile(filename)
	f := peer.ReadFile(filename)
	conn := peer.GetConnection(node.Ip + ":" + node.Port)
	var file_id *uint32
	conn.Call("Peer.SetFile", f, file_id)
	conn.Close()
	serverMessage(filename + " stored successfully.")
}

func retrieve(filename string) {
	node, err := searchFile(filename)
	if err != nil {
		println(">​Server Response:", err)
	} else {
		conn := peer.GetConnection(node.Ip + ":" + node.Port)
		file_id := peer.Hash(filename)
		f := &peer.File{}
		err = conn.Call("Peer.GetFile", file_id, f)
		if err != nil {
			serverMessage("Error:" + err.Error())
		}
		conn.Close()
		peer.WriteFile(f)
		serverMessage(filename + " found.")
	}
}

func handleOptions() {
	for {
		fmt.Print(">Please select an option: ")
		option := readline()

		switch option {
		case "1":
			fmt.Print(">Enter the filename to store: ")
			filename := readline()
			store(filename)
		case "2":
			fmt.Print(">Enter the filename to retrieve: ")
			filename := readline()
			retrieve(filename)
		case "3":
			fmt.Println("Exiting")
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func main() {
	ip = os.Args[1]
	port = os.Args[2]

	fmt.Println(`
	1) Enter the filename to store: 
	2) Enter the filename to retrieve:
	3) Exit:`)

	handleOptions()
}
