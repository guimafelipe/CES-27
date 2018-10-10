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
var ServConn *net.UDPConn
var CliConn map[int]*net.UDPConn

var id int

var latest []int
var engager []int
var num []int
var wait []bool

var neightbours []int

var state int // 0: Released, 1: Wanted, 2: Using

type Message struct{
	Tipo string // Can be Q or R
	I int
	M int
	K int
	J int
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
		if state == 1 {
			err = json.Unmarshal(buf[:n], &msg)
			tipo := msg.Tipo
			m := msg.M
			i := msg.I
			j := msg.K
			k := msg.J
			fmt.Printf("Recebi mensagem: %s(%v, %v, %v, %v)\n", tipo, i, m, j, k)
			if(tipo == "Q"){ //query type message
				if m > latest[i] { //engaging query
					latest[i] = m; engager[i] = j
					wait[i] = true; num[i] = len(neightbours)
					for r := 0; r < len(neightbours); r++ {
						go sendMes("Q", i, m, k, neightbours[r])
					}
				} else if wait[i] && (m == latest[i]){ //not engaging query
					go sendMes("R", i, m, k, j)
				}
			} else if (tipo == "R") {
				if wait[i] && (m == latest[i]) {
					num[i]--
					if num[i] == 0 {
						if i == k {
							fmt.Println("DEADLOCK!!!!")
						} else {
							fmt.Println(engager)
							go sendMes("R", i, m, k, engager[i])
						}
					}
				}
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

	ServConn = ServerConn
}

func main() {
	initConnections()

	state = 0

	defer ServConn.Close()
	for i := 1; i < nServers; i++ {
		if i == id {continue}
		defer CliConn[i].Close()
	}

	wait = make([]bool, nServers+1)
	engager = make([]int, nServers+1)
	num = make([]int, nServers+1)
	latest = make([]int, nServers+1)
	state = 1

	ch := make(chan string) 
	go readInput(ch)

	for {
		go doServerJob()
		select {
		case msg, valid:= <-ch:
			if valid{
				if msg[0:1] == "a" {
					new_nbh,_ := strconv.Atoi(msg[2:])
					neightbours = append(neightbours, new_nbh)
					fmt.Println(neightbours)
				} else if msg == "start" {
					go startAlg()
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

func startAlg(){
	fmt.Println("Starting Algorithm")
	state = 1 //Wanted
	time.Sleep(time.Second*2)
	latest[id]++
	m:=latest[id]
	wait[id]=true
	num[id]=len(neightbours)

	fmt.Println(num[id])

	for i := 0; i < len(neightbours); i++ {
		go sendMes("Q", id, m, id, neightbours[i])
	}
}

func sendMes(tipo string, i int, m int, k int, j int){
		var msgobj Message = Message{tipo, i, m, k, j}
		msg, _ := json.Marshal(msgobj)
		go doClientJob(j, msg)
		fmt.Printf("Envidando mensagem: %s(%v, %v, %v, %v)\n", tipo, i, m, k, j)
}
