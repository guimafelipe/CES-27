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
var nServers int
var CliConn []*net.UDPConn
var ServConn *net.UDPConn
var id int

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
		n,addr,err := ServConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error: ",err)
		}
		fmt.Println("Received ",string(buf[0:n]), " from ",addr)
		var m Clocks
		err = json.Unmarshal(buf[:n], &m)
		if err != nil {
			fmt.Println("Unmarshal Error: ",err)
		}

		fmt.Println("Received data from ",addr)
		for i := 0; i < nServers;i++{
			if myclocks.Clocks[i] < m.Clocks[i]{
				myclocks.Clocks[i] = m.Clocks[i]
			}
		}
		myclocks.Clocks[id-1]++
	}
}

func doClientJob(j int, msg []byte) {
	_,err := CliConn[j].Write(msg)
	if err != nil {
		fmt.Println(string(msg), err)
	}
	time.Sleep(time.Second * 1)
}

func initConnections(){
	id,_ = strconv.Atoi(os.Args[1])
	nServers = len(os.Args) - 2

	for i := 0; i < nServers; i++ {
		if(i+1 == id){
			myPort = os.Args[i+2]
			continue
		}
		clientPort := os.Args[i+2]		
		fmt.Println(clientPort)

		clientAddr,err := net.ResolveUDPAddr("udp","127.0.0.1" + clientPort)
		CheckError(err)

		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1" + ":0")
		CheckError(err)

		Conn, err := net.DialUDP("udp", LocalAddr, clientAddr)
		CliConn = append(CliConn, Conn)
		CheckError(err)
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
	for i := 0; i < nServers-1; i++ {
		defer CliConn[i].Close()
	}

	ch := make(chan string) 
	go readInput(ch)

	for {
		go doServerJob()
		select {
		case x, valid:= <-ch:
			if valid{
				dest, _ := strconv.Atoi(x)
				myclocks.Clocks[id-1]++
				if dest != id {
					sendMessage(dest)
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

func sendMessage(dest int){
	if dest > id{ dest--}
	dest--
	msg, err := json.Marshal(myclocks)
	CheckError(err)
	doClientJob(dest, msg)
}