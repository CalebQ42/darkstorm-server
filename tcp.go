package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func tcpLinker() {
	links, err := parseConf()
	if links == nil {
		log.Println("No values in config file or file not present (/etc/darkstorm-server.conf). tcp linker signing off")
		quitChan <- "tcp conf"
		return
	} else if err != nil {
		log.Println("Error while trying to parse config file:", err)
		quitChan <- "tcp conf"
		return
	}
	fails := make(map[int]int) //logs how many fails per 5 seconds
	failChan := make(chan int, 20)
	open := make(map[int]bool)
	for port, addr := range links {
		open[port] = true
		go link(port, addr, failChan)
	}
failWaiting:
	for portFail := <-failChan; ; portFail = <-failChan {
		if fails[portFail] == 0 {
			go func() {
				time.Sleep(5 * time.Second)
				fails[portFail] = 0
			}()
		} else if fails[portFail] == 4 {
			log.Println("Port", portFail, "has failed 5 time is as many seconds. Not restarting port...")
			open[portFail] = false
			for _, b := range open {
				if b {
					continue failWaiting
				}
			}
			log.Println("All ports dead. Attempting restart...")
			quitChan <- "tcp err"
			return
		}
		fails[portFail]++
		log.Println("Restarting linking for port", portFail)
		go link(portFail, links[portFail], failChan)
	}

}

func link(port int, addr string, failChan chan int) {
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	defer listen.Close()
	if err != nil {
		log.Println("Error while trying to listen to port ", port, ":", err)
		failChan <- port
		return
	}
	for {
		con, err := listen.Accept()
		defer con.Close()
		if err != nil {
			log.Println("Error while trying to accept connection to port ", port, ":", err)
			failChan <- port
			return
		}
		ext, err := net.Dial("tcp", addr)
		defer ext.Close()
		if err != nil {
			log.Println("Error while trying to dial", addr, ":", err)
			failChan <- port
			return
		}
		_, err = io.Copy(ext, con)
		if err != nil {
			log.Println("Error while trying to copy data to", addr, ":", err)
			failChan <- port
			return
		}
		_, err = io.Copy(con, ext)
		if err != nil {
			log.Println("Error while trying to copy data to port", port, ":", err)
			failChan <- port
			return
		}
	}
}

func parseConf() (links map[int]string, err error) {
	conf, err := os.Open("/etc/darkstorm-server.conf")
	if os.IsNotExist(err) {
		return nil, nil
	}
	lineNum := 0
	links = make(map[int]string)
	rdr := bufio.NewReader(conf)
	for {
		lineNum++
		var origLine string
		origLine, err = rdr.ReadString('\n')
		if err != nil {
			break
		}
		line := strings.ReplaceAll(origLine, "\t", " ")
		for strings.Contains(line, "  ") {
			line = strings.Replace(line, "  ", " ", -1)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		split := strings.Split(line, " ")
		if len(split) != 2 {
			return nil, errors.New("invalid line #" + strconv.Itoa(lineNum))
		}
		var i int
		i, err = strconv.Atoi(split[0])
		if err != nil {
			return nil, errors.New("invalid line #" + strconv.Itoa(lineNum))
		}
		links[i] = split[1]
	}
	err = nil
	if len(links) > 0 {
		return nil, nil
	}
	return
}
