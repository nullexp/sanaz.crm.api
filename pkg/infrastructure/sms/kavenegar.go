package sms

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func NewKavenegarSMSNotifier(api string, timeout time.Duration, keepAlive time.Duration) SMSNotifier {
	return &KavenegarSMSNotifier{
		api: api,
		httpTransport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: keepAlive,
			}).Dial,
			TLSHandshakeTimeout: 60 * time.Second,
			TLSClientConfig: &tls.Config{
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.CurveP256,
					tls.X25519,
				},
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
		},
	}
}

type KavenegarSMSNotifier struct {
	api           string
	httpTransport *http.Transport
}

func (k KavenegarSMSNotifier) SendMessage(req Message) (err error) {
	return sendSMS(k.api, k.httpTransport, req)
}

func (k KavenegarSMSNotifier) SendBulkMessages(messages []Message) error {
	for _, v := range messages {
		if err := sendSMS(k.api, k.httpTransport, v); err != nil {
			return err
		}
	}

	return nil
}

func (k KavenegarSMSNotifier) SendBulkMessagesAsync(messages []Message) <-chan error {
	errs := make(chan error)
	go func() {
		defer close(errs)
		for _, v := range messages {
			if err := sendSMS(k.api, k.httpTransport, v); err != nil {
				errs <- err
			}
		}
	}()
	return errs
}

func sendSMS(api string, httpTransport *http.Transport, message Message) error {
	url := fmt.Sprintf(
		"%s?receptor=%s&token=%s&template=%s",
		api,
		message.PhoneNumber,
		strings.ReplaceAll(message.Text, " ", "_"),
		"InviteUserToOrganization",
	)
	if err := send(url, httpTransport); err != nil {
		log.Printf("the following error occurred while sending message:%s \n", err)
		return err
	}

	return nil
}

func send(url string, httpTransport *http.Transport) error {
	client := &http.Client{Transport: httpTransport}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("the following error occurred while sending message:%s \n", err)
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("send", "error", err)
		return err
	}
	if res.StatusCode != http.StatusOK {
		log.Println("Response status code not OK", "status code", res.StatusCode)
		return fmt.Errorf("failed to send SMS")
	}
	return nil
}
