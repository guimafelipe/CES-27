package main

import (
	"fmt"
	"net"
	"os"
	// "time"
	// "strconv"
	// "bufio"
	"encoding/json"
)

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

func main(){
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)
	Connection, err := net.ListenUDP("udp", Address)
	CheckError(err)
	defer Connection.Close()
	buf := make([]byte, 1024)
	for {
		n,_,err := Connection.ReadFromUDP(buf)
		if err!= nil{
			fmt.Println("Error: ", err)
		}
		var msg Message
		err = json.Unmarshal(buf[:n], &msg)
		fmt.Println("Processo ",msg.Id, " me usando agora, com clock ",msg.Clock)
	}
}