package main

import (
	"flag"
	"log"
	"time"
)

var quitChan chan string = make(chan string)

func main() {
	mongoStr := flag.String("mongo", "", "MongoDB connection string for APIs")
	flag.Parse()
	go linker()
	go webserver(*mongoStr)
	for failure := <-quitChan; ; failure = <-quitChan {
		switch failure {
		case "tcp conf":
			continue
		case "tcp err":
			go tcpLinkerRestart()
		case "web arg":
			continue
		case "web err":
			go websiteRestart(*mongoStr)
		}
	}
}

func tcpLinkerRestart() {
	log.Println("TCP linker failed. Restarting in 5 seconds...")
	time.Sleep(5 * time.Second)
	log.Println("Restarting tcp linker")
	linker()
}

func websiteRestart(mongoStr string) {
	log.Println("Website failed. Restarting in 5 seconds...")
	time.Sleep(5 * time.Second)
	log.Println("Restarting website")
	webserver(mongoStr)
}