package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
var current_username string = ""

type P struct {
	Rtype   string `json:"rtype"`
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

func send(p P) (P, net.Conn, error) {
	conn, err := net.Dial("tcp", ip+":"+port)
	encoder := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	if err != nil {
		fmt.Println("Couldn't establish connection on", ip+":"+port)
		return P{}, conn, err
	}
	encoder.Encode(p)
	var response P
	err = dec.Decode(&response)

	if err != nil {
		fmt.Println("Error during decoding message from the server", err)
		return P{}, conn, err
	}

	return response, conn, nil
}

func login(username string) {
	request := P{"L", "", username}
	response, conn, err := send(request)
	if err != nil || response.Rtype != "L" {
		fmt.Println("Couldn't login ", err)
		serverMessage(response.Message)
		return
	}

	serverMessage(response.Message)
	current_username = username
	conn.Close()
}

func store(filename string) {
	if current_username == "" {
		fmt.Println("You need to login first!")
		return
	}

	request := P{"S", filename, current_username}
	response, conn, err := send(request)
	if err != nil {
		fmt.Println("Couldn't store file", err)
		return
	} else if response.Rtype != "S" {
		serverMessage(response.Message)
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Couldn't read file", err)
		return
	}

	defer f.Close()
	_, err = io.Copy(conn, f)
	if err != nil {
		fmt.Println("Couldn't read file", err)
		return
	}

	serverMessage(response.Message)
	conn.Close()
}

func retrieve(filename string) {
	if current_username == "" {
		fmt.Println("You need to login first!")
		return
	}

	request := P{"R", filename, current_username}
	response, conn, err := send(request)
	if err != nil {
		fmt.Println("Couldn't retrieve file", err)
		return
	} else if response.Rtype != "R" {
		serverMessage(response.Message)
		return
	}

	if _, err := os.Stat(current_username); os.IsNotExist(err) {
		os.Mkdir(current_username, os.ModePerm)
	}

	file, err := os.Create(current_username + "/" + filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()
	_, err = io.Copy(file, conn)
	if err != nil {
		log.Fatal(err)
	}
	serverMessage(response.Message)
	conn.Close()
}

func handleOptions() {
	for {
		fmt.Print(">Please select an option: ")
		option := readline()

		switch option {
		case "1":
			fmt.Print(">Enter the username: ")
			username := readline()
			login(username)
		case "2":
			fmt.Print(">Enter the filename to store: ")
			filename := readline()
			store(filename)
		case "3":
			fmt.Print(">Enter the filename to retrieve: ")
			filename := readline()
			retrieve(filename)
		case "4":
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
	1) Enter the username:
	2) Enter the filename to store: 
	3) Enter the filename to retrieve:
	4) Exit:`)

	handleOptions()
}
