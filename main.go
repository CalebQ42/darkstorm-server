package main

import (
	"crypto/tls"
	"flag"
	"log"
)

func main() {
	mongoURL := flag.String("mongo", "", "Enables MongoDB usage for darkstorm-backend.")
	webRoot := flag.String("web-root", "", "Sets root directory of web server.")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("You must specify key directory. ex: darkstorm-server /etc/web-keys")
	}
	if *mongoURL != "" {
	}
	mongoCert, err := tls.LoadX509KeyPair(flag.Arg(0)+"mongo.pem", flag.Arg(0)+"key.pem")
	if err != nil {

	}
}
