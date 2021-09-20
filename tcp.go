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
	if err != nil {
		log.Println("Error while trying to parse config file:", err, "tcp linker signing off")
		quitChan <- "tcp conf"
		return
	} else if links == nil {
		log.Println("No values in config file or file not present (/etc/darkstorm-server.conf). tcp linker signing off")
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
	if err != nil {
		log.Println("Error while trying to listen to port ", port, ":", err)
		failChan <- port
		return
	}
	defer listen.Close()
	for {
		con, err := listen.Accept()
		if err != nil {
			log.Println("Error while trying to accept connection to port ", port, ":", err)
			failChan <- port
			return
		}
		err = copyConn(con, addr)
		if err != nil {
			log.Println("Error while trying copy data from port", port, "to address", addr, ":", err)
			failChan <- port
			return
		}
	}
}

func parseConf() (links map[int]string, err error) {
	conf, err := os.Open("/etc/darkstorm-server.conf")
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	lineNum := 0
	links = make(map[int]string)
	rdr := bufio.NewReader(conf)
	multilineComment := false
	var line string
	for {
		if line == "" {
			lineNum++
			line, err = rdr.ReadString('\n')
			if err != nil && line == "" {
				break
			} else if line == "" {
				continue
			}
		}
		startCom, endCom := strings.Index(line, "/*"), strings.Index(line, "*/")
		if multilineComment {
			if endCom != -1 {
				line = line[endCom:]
			} else {
				continue
			}
		}
		if startCom != -1 {
			if endCom != -1 {
				line = line[:startCom] + line[endCom:]
				continue
			}
			line = line[:startCom]
			multilineComment = true
		}
		if strings.Contains(line, "//") {
			line = line[:strings.Index(line, "//")]
		}
		line = strings.ReplaceAll(line, "\t", " ")
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
		line = ""
	}
	err = nil
	if len(links) == 0 {
		return nil, nil
	}
	return
}

func copyConn(src net.Conn, addr string) error {
	dst, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Erro while dialing", addr)
		return err
	}

	done := make(chan struct{})

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
	<-done
	return nil
}
