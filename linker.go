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

type link struct {
	addr     string
	linkType string
}

func (l link) isTCP() bool {
	return strings.HasPrefix(l.linkType, "tcp") || strings.HasPrefix(l.linkType, "unix")
}

func (l link) isUDP() bool {
	return strings.HasPrefix(l.linkType, "udp")
}

func linker() {
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
		go createLink(port, addr, failChan)
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
		go createLink(portFail, links[portFail], failChan)
	}

}

func createLink(port int, l link, failChan chan int) {
	log.Println("Linking", port, "to", l.addr, "with type", l.linkType)
	var tcpListen net.Listener
	var con net.Conn
	var err error
	if l.isTCP() {
		tcpListen, err = net.Listen(l.linkType, ":"+strconv.Itoa(port))
		if err != nil {
			log.Println("Error while trying to listen to port", port, ":", err)
			failChan <- port
			return
		}
		defer tcpListen.Close()
	} else if l.isUDP() {
		var addr *net.UDPAddr
		addr, err = net.ResolveUDPAddr(l.linkType, ":"+strconv.Itoa(port))
		if err != nil {
			log.Println("Error while parsing port", port, ":", err)
			failChan <- port
			return
		}
		con, err = net.ListenUDP(l.linkType, addr)
		if err != nil {
			log.Println("Error while listening to port", port, ":", err)
		}
	}
	for {
		if l.isTCP() {
			con, err = tcpListen.Accept()
		}
		if err != nil {
			log.Println("Error while trying to accept connection to port ", port, ":", err)
			failChan <- port
			return
		}
		err = copyConn(con, l)
		if err != nil {
			log.Println("Error while trying copy data from port", port, "to address", l.addr, ":", err)
			failChan <- port
			return
		}
		if l.isUDP() {
			var addr *net.UDPAddr
			addr, err = net.ResolveUDPAddr(l.linkType, ":"+strconv.Itoa(port))
			if err != nil {
				log.Println("Error while parsing port", port, ":", err)
				failChan <- port
				return
			}
			con, err = net.ListenUDP(l.linkType, addr)
			if err != nil {
				log.Println("Error while listening to port", port, ":", err)
			}
		}
	}
}

func parseConf() (links map[int]link, err error) {
	conf, err := os.Open("/etc/darkstorm-server.conf")
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	lineNum := 0
	links = make(map[int]link)
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
		if len(split) < 2 || len(split) > 3 {
			return nil, errors.New("invalid line #" + strconv.Itoa(lineNum))
		}
		var l link
		if len(split) == 3 {
			l.linkType = split[0]
			split = split[1:]
		} else {
			l.linkType = "tcp"
		}
		var i int
		i, err = strconv.Atoi(split[0])
		if err != nil {
			return nil, errors.New("invalid line #" + strconv.Itoa(lineNum))
		}
		l.addr = split[1]
		links[i] = l
		line = ""
	}
	err = nil
	if len(links) == 0 {
		return nil, nil
	}
	return
}

func copyConn(src net.Conn, l link) error {
	dst, err := net.Dial(l.linkType, l.addr)
	if err != nil {
		log.Println("Error while dialing", l.addr)
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
