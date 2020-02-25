package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type P struct {
	Rtype   string `json:"rtype"`
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

func send(p P, conn net.Conn) {
	err := json.NewEncoder(conn).Encode(p)
	if err != nil {
		fmt.Println("Error during encoding message", err)
		conn.Close()
	}
}

func login(p P, conn net.Conn) {
	response := P{"L", "Login Successful", ""}
	send(response, conn)
	conn.Close()
}

func store(p P, conn net.Conn) {

	if _, err := os.Stat(p.Sender); os.IsNotExist(err) {
		os.Mkdir(p.Sender, os.ModePerm)
	}

	file, err := os.Create(p.Sender + "/" + p.Message)
	response := P{"S", p.Message + " stored successfully", ""}
	if err != nil {
		fmt.Println(err)
		response.Message = "Error opening file: " + p.Message
		send(response, conn)
		return
	}

	send(response, conn)
	defer file.Close()
	_, err = io.Copy(file, conn)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()
}

func retrieve(p P, conn net.Conn) {
	f, err := os.Open(p.Sender + "/" + p.Message)
	if err != nil {
		fmt.Println(err)
		p.Message = "​File does not exist"
		p.Rtype = "E"
		send(p, conn)
		return
	}

	response := P{"R", p.Message + " found.", ""}
	send(response, conn)

	defer f.Close()
	_, err = io.Copy(conn, f)
	if err != nil {
		fmt.Println("Couldn't send file", f, err)
	}
	conn.Close()
}

func handleConnection(conn net.Conn) {

	dec := json.NewDecoder(conn)

	request := P{}
	err := dec.Decode(&request)

	if err != nil {
		fmt.Println("Error during decoding message", err)
		encoder := json.NewEncoder(conn)
		encoder.Encode(P{"E", "Error during decoding message" + err.Error(), ""})
		conn.Close()
		return
	}

	switch request.Rtype {
	case "L":
		login(request, conn)
	case "S":
		store(request, conn)
	case "R":
		retrieve(request, conn)
	default:
		response := &P{}
		response.Message = "Bad request."
		encoder := json.NewEncoder(conn)
		encoder.Encode(*response)
		conn.Close()
	}
}

func main() {
	var port = os.Args[1]
	fmt.Println("Start listening on port", port)
	ln, err := net.Listen("tcp", "localhost:"+port)

	if err != nil {
		fmt.Println("Couldn't start the server", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Couldn't establish tcp connection, error", err)
		}
		go handleConnection(conn)
	}
}
