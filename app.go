package main

import (
	"os"
	"net"
	"fmt"
	"bufio"
	"strings"
	"encoding/json"
)

type Node struct {
	Connections map[string]bool 
	Address Address
}

type Address struct {
	IPv4 string
	Port string
}

type Package struct {
	To string
	From string
	Data string
}

func init(){
	if len(os.Args) != 2 {panic("Enter IPv4:PORT")}
}

func main(){
	Start()
}

func Start(){
	NewNode(os.Args[1]).Run(handleServer, handleClient)
}

func NewNode(address string) *Node {
	splited := strings.Split(address, ":")
	if len(splited) != 2 {return nil}
	return &Node{
		Connections: make(map[string]bool),
		Address: Address{
			IPv4: splited[0],
			Port: ":" + splited[1],
		},
	}
}

func (node *Node) Run(handleServer func(*Node), handleClient func(*Node)){
	go handleServer(node)
	handleClient(node)
}

func handleServer(node *Node) {
	listen, err := net.Listen("tcp", node.Address.IPv4 + node.Address.Port)
	if err != nil {
		fmt.Println(err)
	}
	defer listen.Close()
	
	for {
		conn, err := listen.Accept()
		if err != nil {break}
		go handleConnection(node, conn)
	}
}

func handleConnection(node *Node, conn net.Conn) {
	defer conn.Close()
	var (
		buffer = make([]byte, 512)
		message string
		pack Package
	)
	for {
		length, err := conn.Read(buffer)
		if err != nil {break}
		message += string(buffer[:length])
	}
	err := json.Unmarshal([]byte(message), &pack)
	if err != nil {
		return
	}
	node.ConnectTo([]string{pack.From})
	fmt.Println(pack.Data)
}

func handleClient(node *Node) {
	for {
		message := InputString()
		splited := strings.Split(message, " ")
		switch splited[0]{
			case "#exit": os.Exit(0)
			case "#network": node.PrintNetwork()
			case "#connect": node.ConnectTo(splited[0:])
			default: node.SendMessageToAll(node.Address.IPv4+node.Address.Port+ " >>> " + message)
		}
	}	
}

func (node *Node) PrintNetwork() {
	for addr := range node.Connections {
		fmt.Println(addr, "\n")
	}
}

func (node *Node) ConnectTo(addresses []string) {
	for _, addr := range addresses {
		node.Connections[addr] = true
	}
}

func (node *Node) SendMessageToAll(message string){
	var new_pack = Package{
		From: node.Address.IPv4 + node.Address.Port,
		Data: message,
	}
	for addr := range node.Connections {
		new_pack.To = addr
		node.Send(&new_pack)
	}
}

func (node *Node) Send(pack *Package) {
	conn, err := net.Dial("tcp", pack.To)
	if err != nil {
		delete(node.Connections, pack.To)
		return
	}
	defer conn.Close()
	json_pack, _  := json.Marshal(*pack)
	conn.Write(json_pack)
}

func InputString() string {
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", -1)
}
