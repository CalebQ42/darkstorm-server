package main

import (
	"log"
	"time"
)

var quitChan chan string = make(chan string)

func main() {
	go tcpLinker()
	go webserver()
	for failure := <-quitChan; ; failure = <-quitChan {
		switch failure {
		case "tcp conf":
			continue
		case "tcp err":
			go tcpLinkerRestart()
		case "web arg":
			continue
		case "web err":
			go websiteRestart()
		}
	}
}

func tcpLinkerRestart() {
	log.Println("TCP linker failed. Restarting in 5 seconds...")
	time.Sleep(5 * time.Second)
	log.Println("Restarting tcp linker")
	tcpLinker()
}

func websiteRestart() {
	log.Println("Website failed. Restarting in 5 seconds...")
	time.Sleep(5 * time.Second)
	log.Println("Restarting website")
	webserver()
}
