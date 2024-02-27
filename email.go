package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"time"

	smtp "github.com/emersion/go-smtp"
)

func startSMTPServer() {
	keyPath := flag.Arg(1)
	if keyPath == "" {
		log.Println("No argument given for key files. smtp signing off...")
		quitChan <- "smtp arg"
		return
	}
	tlsCert, err := tls.LoadX509KeyPair(keyPath+"/fullchain.pem", keyPath+"/key.pem")
	if err != nil {
		log.Println("Error while loading tls certificate:", err)
		quitChan <- "smtp err"
		return
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	serv := smtp.NewServer(&SMTPBackend{
		tlsConfig: cfg,
	})
	serv.Network = "tcp"
	serv.Addr = ":587"
	serv.TLSConfig = cfg
	serv.Domain = "darkstorm.tech"
	serv.WriteTimeout = 10 * time.Second
	serv.ReadTimeout = 10 * time.Second
	serv.MaxMessageBytes = 1024 * 1024
	serv.AllowInsecureAuth = true
	serv.MaxRecipients = 2
	err = serv.ListenAndServeTLS()
	log.Println("Error while serving smtp:", err)
	quitChan <- "smtp err"
}

type SMTPBackend struct {
	tlsConfig *tls.Config
}

func (b *SMTPBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return NewSession(b.tlsConfig, c.Server().Addr)
}

type SMTPSession struct {
	tlsConfig *tls.Config
	client    *smtp.Client
	addr      string
}

func NewSession(tlsConfig *tls.Config, addr string) (*SMTPSession, error) {
	client, err := smtp.DialTLS(addr, tlsConfig)
	if err != nil {
		return nil, err
	}
	return &SMTPSession{
		tlsConfig: tlsConfig,
		client:    client,
		addr:      addr,
	}, nil
}

func (s *SMTPSession) Reset() {
	var err error
	s.client, err = smtp.DialTLS(s.addr, s.tlsConfig)
	if err != nil {
		log.Println("Error while resetting smtp session:", err)
	}
}

func (s *SMTPSession) Logout() error {
	return s.client.Quit()
}

func (s *SMTPSession) AuthPlain(username, password string) error {
	//TODO
	return nil
}

func (s *SMTPSession) Mail(from string, opts *smtp.MailOptions) error {
	return s.client.Mail(from, opts)
}

func (s *SMTPSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	return s.client.Rcpt(to, opts)
}

func (s *SMTPSession) Data(r io.Reader) error {
	wrt, err := s.client.Data()
	if err != nil {
		log.Println("Error while writing smtp data:", err)
		return err
	}
	_, err = io.Copy(wrt, r)
	if err != nil {
		log.Println("Error while writing smtp data:", err)
		return err
	}
	wrt.Close()
	return nil
}
