package main

import (
	"fmt"
	"net"
	"os"
	"time"
	"strconv"
)

var err string
var myPort string
var nServers int
var CliConn []*net.UDPConn
var ServConn *net.UDPConn

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

	// defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n,addr,err := ServConn.ReadFromUDP(buf)
		fmt.Println("Received ",string(buf[0:n]), " from ",addr)

		if err != nil {
			fmt.Println("Error: ",err)
		}
	}
}

func doClientJob(j int, i int) {
	// defer Conn.Close()
	msg := strconv.Itoa(i)
	i++
	buf := []byte(msg)
	_,err := CliConn[j].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}
	time.Sleep(time.Second * 1)
}

func initConnections(){
	myPort = os.Args[1]
	nServers = len(os.Args) - 2

	ServerAddr,err := net.ResolveUDPAddr("udp", myPort)
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	ServConn = ServerConn

	for i := 0; i < nServers; i++ {
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

}

func main() {
	initConnections()

	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	i := 0
	for {
		go doServerJob()
		for j := 0; j < nServers; j++ {
			go doClientJob(j, i)
		}
		time.Sleep(time.Second * 1)
		i++
	}
}