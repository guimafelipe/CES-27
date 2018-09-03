package main

import (
	"fmt"
	"net"
	"os"
	"time"
	"strconv"
	"bufio"
)

var err string
var myPort string
var nServers int
// var CliConn []*net.UDPConn
var ServConn *net.UDPConn
var id int
var next int
var leader int
var CliConn map[int]*net.UDPConn
var isParticipant bool = false

type Clocks struct{
	Id int
	Clocks [3]int
}

var myclocks Clocks

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

		msg := string(buf[0:n])
		fmt.Printf("P%v: Recebeu %s\n", id, msg)

		candidate,_ := strconv.Atoi(msg[1:])

		if(msg[0] == 'S'){
			if candidate > id {
				isParticipant = true
				doClientJob(next, msg)
			} else if candidate < id {
				if !isParticipant {
					isParticipant = true
					sendMessage(next, "S")
				}
			} else {
				//Sou o lider
				fmt.Println("Sou o lider")
				isParticipant = false
				sendMessage(next, "F")
			}
		} else if (msg[0] == 'F' && isParticipant){
			isParticipant = false
			leader = candidate
			doClientJob(next, msg)
		}
	}
}

func doClientJob(j int, msg string) {
	time.Sleep(time.Second*2)
	buf := []byte(msg)	
	_,err := CliConn[j].Write(buf)
	if err != nil {
		fmt.Println(string(msg), err)
	}
	fmt.Printf("P%v: Enviou %s\n", id, msg)
	time.Sleep(time.Second * 1)
}

func initConnections(){
	id,_ = strconv.Atoi(os.Args[1])
	next,_ = strconv.Atoi(os.Args[2])
	nServers = len(os.Args) - 3
	CliConn = make(map[int]*net.UDPConn)

	for i := 0; i < nServers; i++ {
		if(i+1 == id){
			myPort = os.Args[i+3]
			continue
		}

		clientPort := os.Args[i+3]		

		clientAddr,err := net.ResolveUDPAddr("udp","127.0.0.1" + clientPort)
		CheckError(err)

		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1" + ":0")
		CheckError(err)

		Conn, err := net.DialUDP("udp", LocalAddr, clientAddr)
		CheckError(err)

		CliConn[i+1] = Conn
	}

	ServerAddr,err := net.ResolveUDPAddr("udp", myPort)
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	ServConn = ServerConn
}

func main() {
	initConnections()

	myclocks = Clocks{id, [3]int{0,0,0}}

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
				if msg == "start" {
					sendMessage(next, "S")
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
	var msg string
	msg = tipo + strconv.Itoa(id)
	doClientJob(dest, msg)
}