package main

import (
	"fmt"
	"net"
	"os"
	"time"
	"strconv"
	"bufio"
	"encoding/json"
)

var err string
var myPort string
var sharedResPort string = ":10001"
var nServers int
// var CliConn []*net.UDPConn
var ServConn *net.UDPConn
var ResConn *net.UDPConn
var id int
var next int
var leader int
var CliConn map[int]*net.UDPConn
var isParticipant bool = false
var myclock int
var nOks int
var queue []int
var state int // 0: Released, 1: Wanted, 2: Using

type Message struct{
	Id int
	Clock int
}

func CheckError(err error){
	if err != nil{
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}

func PrintError(err error){
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

func doServerJob(){
	buf := make([]byte, 1024)

	for {
		n,_,err := ServConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error: ",err)
		}

		var msg Message
		var smsg string
		smsg = string(buf[:n])
		if smsg == "OK"{
			nOks++
		} else {
			fmt.Println("Mensagem chegou, estado: ", state)
			err = json.Unmarshal(buf[:n], &msg)
			if state == 0 {
				go sendOk(msg.Id)
			} else if state == 1{
				if msg.Clock < myclock {
					go sendOk(msg.Id)
				} else {
					queue = append(queue, msg.Id)
				}
			} else if state == 2{
				queue = append(queue, msg.Id)
			}
		}
	}
}

func doClientJob(j int, msg []byte) {
	_,err := CliConn[j].Write(msg)
	if err != nil {
		fmt.Println(string(msg), err)
	}
}

func initConnections(){
	id,_ = strconv.Atoi(os.Args[1])
	nServers = len(os.Args) - 2
	CliConn = make(map[int]*net.UDPConn)

	for i := 1; i <= nServers; i++ {
		if(i == id){
			myPort = os.Args[i+1]
			continue
		}

		clientPort := os.Args[i+1]		
		// fmt.Println("Client port:", clientPort)

		clientAddr,err := net.ResolveUDPAddr("udp","127.0.0.1" + clientPort)
		CheckError(err)

		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1" + ":0")
		CheckError(err)

		Conn, err := net.DialUDP("udp", LocalAddr, clientAddr)
		CheckError(err)

		CliConn[i] = Conn
	}

	ServerAddr,err := net.ResolveUDPAddr("udp", myPort)
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	ResourceAddr, err := net.ResolveUDPAddr("udp", sharedResPort)
	CheckError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1" + ":0")
	CheckError(err)
	ResourceConn, err := net.DialUDP("udp", LocalAddr, ResourceAddr)
	CheckError(err)

	ServConn = ServerConn
	ResConn = ResourceConn
}

func main() {
	initConnections()

	// myclocks = Clocks{id, [3]int{0,0,0}}
	myclock = 0
	state = 0
	queue = make([]int, 0)

	defer ServConn.Close()
	for i := 1; i < nServers; i++ {
		if i == id {continue}
		defer CliConn[i].Close()
	}

	ch := make(chan string) 
	go readInput(ch)

	for {
		go doServerJob()
		select {
		case msg, valid:= <-ch:
			if valid{
				imsg, _ := strconv.Atoi(msg)
				if msg == "x" {
					// sendMessage(next, "S")
					go askForResource()
				} else if imsg == id {
					myclock+=1
				}
			} else {
				fmt.Println("Channel closed!")
			}
		default:
			time.Sleep(time.Second*1)
		}
		time.Sleep(time.Second*1)
	}
}

func readInput(ch chan string){
	reader:= bufio.NewReader(os.Stdin)
	for {
		text, _, _:= reader.ReadLine()
		ch <- string(text)
	}
}

func sendMessage(dest int, tipo string){
	var msgobj Message
	msg, _ := json.Marshal(msgobj)
	doClientJob(dest, msg)
}

func askForResource(){
	state = 1 //Wanted
	time.Sleep(time.Second*2)
	var msgobj Message = Message{id, myclock}
	msg, _ := json.Marshal(msgobj)
	nOks = 0
	for i := 1; i <= nServers; i++ {
		if i == id{continue}
		go doClientJob(i, msg)
	}

	for nOks < nServers - 1 {

	}

	fmt.Println("Usando a CS")
	state = 2
	//Using
	_,err := ResConn.Write(msg)
	time.Sleep(time.Second*2)
	fmt.Println("Terminei de usar a CS")

	state = 0
	//Releasing
	for i := 0; i < len(queue); i++ {
		go sendOk(queue[i])
	}

	queue = make([]int, 0)

	if err != nil {
		fmt.Println(string(msg), err)
	}
}

func sendOk(clid int){
	buf := []byte("OK")
	CliConn[clid].Write(buf)
	fmt.Println("Enviando ok para ", clid)
}