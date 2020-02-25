package peer

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
)

type Peer struct {
	Id   uint32
	Ip   string
	Port string
}

type Setter struct {
	V string
	P Peer
}

type File struct {
	Name    string
	Encoded string
}

var Files = make(map[uint32]string)
var Pred *Peer = nil
var Succ *Peer
var Me *Peer

func Hash(s string) uint32 {
	// https://stackoverflow.com/questions/13582519/how-to-generate-hash-number-of-a-string-in-go#13582881
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func GetOutboundIP() string {
	// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	conn.Close()
	return localAddr.IP.String()
}

func GetConnection(address string) *rpc.Client {
	conn, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return conn
}

func GetSuccessorOf(conn *rpc.Client, id uint32, peer *Peer) error {
	err := conn.Call("Peer.FindSuccessor", id, peer)
	conn.Close()
	return err
}

func GetSuccessor(id uint32) *Peer {
	succ := GetConnection(Succ.Ip + ":" + Succ.Port)
	_peer := &Peer{}
	GetSuccessorOf(succ, id, _peer)
	return _peer
}

func Connect(peeraddress string) {
	// peer x will first search for its successor
	conn := GetConnection(peeraddress)
	// Peer x sets its successor to peer ​y.
	GetSuccessorOf(conn, Me.Id, Succ)
	// Peer x asks peer y for its predecessor and set own predecessor to peer y predecessor.
	conn = GetConnection(Succ.Ip + ":" + Succ.Port)
	err := conn.Call("Peer.Get", "p", Pred)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	conn.Close()
	// Peer x will notify peer y to set its predecessor to x.
	conn = GetConnection(Succ.Ip + ":" + Succ.Port)
	err = conn.Call("Peer.Set", Setter{"p", *Me}, Me)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	conn.Close()
	// Peer x will notify peer y predecessor to set its successor to x.
	conn = GetConnection(Pred.Ip + ":" + Pred.Port)
	err = conn.Call("Peer.Set", Setter{"s", *Me}, Me)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	conn.Close()
}

func (t *Peer) FindSuccessor(id uint32, peer *Peer) error {
	if Pred.Id != Me.Id && id > Pred.Id && id <= Me.Id {
		peer.Id = Me.Id
		peer.Ip = Me.Ip
		peer.Port = Me.Port
		return nil
	} else if (id > Me.Id && id <= Succ.Id) || Succ.Id == Me.Id {
		peer.Id = Succ.Id
		peer.Ip = Succ.Ip
		peer.Port = Succ.Port
		return nil
	} else if Succ.Id != Me.Id {
		succ := GetConnection(Succ.Ip + ":" + Succ.Port)
		GetSuccessorOf(succ, id, peer)
		return nil
	} else {
		return errors.New("Couldn't find peer with the given key")
	}
}

func (t *Peer) Get(p string, reply *Peer) error {
	if p == "s" {
		reply.Id = Succ.Id
		reply.Ip = Succ.Ip
		reply.Port = Succ.Port
	} else if p == "p" {
		reply.Id = Pred.Id
		reply.Ip = Pred.Ip
		reply.Port = Pred.Port
	}
	return nil
}

func (t *Peer) Set(p Setter, reply *Peer) error {
	if p.V == "s" {
		Succ.Id = p.P.Id
		Succ.Ip = p.P.Ip
		Succ.Port = p.P.Port
	} else if p.V == "p" {
		Pred.Id = p.P.Id
		Pred.Ip = p.P.Ip
		Pred.Port = p.P.Port
		// The files from peer y, whose successor is peer x, will move to peer x.
		for k, v := range Files {
			if k >= Pred.Id {
				f := ReadFile(v)
				conn := GetConnection(Pred.Ip + ":" + Pred.Port)
				var file_id *uint32
				_ = conn.Call("Peer.SetFile", f, file_id)
				os.Remove(v)
				delete(Files, k)
				conn.Close()
			}
		}
	}
	return nil
}

func (t *Peer) GetFile(id uint32, reply *File) error {
	if filename, ok := Files[id]; ok {
		f := ReadFile(filename)
		reply.Name = f.Name
		reply.Encoded = f.Encoded
		return nil
	} else {
		return errors.New("​File does not exist.")
	}
}

func (t *Peer) SetFile(file File, id *uint32) error {
	Files[Hash(file.Name)] = file.Name
	WriteFile(&file)
	return nil
}

func (t *Peer) PopFile(id uint32, reply *File) error {
	if filename, ok := Files[id]; ok {
		f := ReadFile(filename)
		reply.Name = f.Name
		reply.Encoded = f.Encoded
		err := os.Remove(f.Name)
		delete(Files, id)
		return err
	} else {
		return errors.New("​File does not exist.")
	}
}

func WriteFile(file *File) {
	data, err := base64.StdEncoding.DecodeString(file.Encoded)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	err = ioutil.WriteFile(file.Name, data, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func ReadFile(filename string) *File {
	f, _ := os.Open(filename)
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)
	encoded := base64.StdEncoding.EncodeToString(content)
	f.Close()
	return &File{Name: filename, Encoded: encoded}
}

func Leave() {
	if Succ.Id != Me.Id {
		for k, v := range Files {
			f := ReadFile(v)
			conn := GetConnection(Succ.Ip + ":" + Succ.Port)
			var file_id *uint32
			_ = conn.Call("Peer.SetFile", f, file_id)
			os.Remove(v)
			delete(Files, k)
			conn.Close()
		}
	}
}
