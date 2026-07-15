package email

import (
	"fmt"
	"log"
	"net/smtp"
)

type Notifier struct {
	host string
	port int
	from string
}

func NewNotifier(host string, port int, from string) *Notifier {
	return &Notifier{host: host, port: port, from: from}
}

func (n *Notifier) SendVetAlert(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", n.from, to, subject, body)
	err := smtp.SendMail(addr, nil, n.from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("email send (dev may use mailhog): %v", err)
	}
	return nil
}
