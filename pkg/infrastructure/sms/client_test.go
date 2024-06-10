//go:build integration

package sms

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/sms/dto"
)

const (
	apiKey    = "676c696f57674e56704b7a66457536705351584a304a6a43355a354c66476f6d6b3958684d56466e5045343d"
	messageID = "1422695019"
)

func TestSendSMS(t *testing.T) {
	sms := NewSMSClient(apiKey)

	request := dto.SendSMSRequest{
		Receptor: []string{"+989333033375"},
		Sender:   []string{"20000110220", "10007119"},
		Message:  "سلام",
	}

	response, err := sms.SendSMS(request)
	if err != nil {
		if err.Error() == "ارسال کننده نامعتبر است" {
			return
		}
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}
}

func TestSendLookUpSMS(t *testing.T) {
	sms := NewSMSClient(apiKey)

	request := dto.LookupSMSRequest{
		Receptor: []string{"+989333033375"},
		Sender:   []string{"20000110220"},
		Message:  "https://espadev.ir/auth",
		Template: "InviteUserToOrganization",
	}

	response, err := sms.Lookup(request)
	if err != nil {
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}
}

func TestStatus(t *testing.T) {
	sms := NewSMSClient(apiKey)

	response, err := sms.StatusByMessageIds([]string{messageID})
	if err != nil {
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}
}

func TestSendBatchSMS(t *testing.T) {
	sms := NewSMSClient(apiKey)

	request := dto.SendBatchSMSRequest{
		Receptor: []string{"09333033375", "09300035055"},
		Sender:   []string{"9876543210", "1234567890"},
		Message:  []string{"Hello, world!"},
	}

	response, err := sms.SendBatchSMS(request)
	if err != nil {
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}

	if len(response.Entries) != 2 {
		t.Errorf("Unexpected number of entries: %d", len(response.Entries))
	}

	if response.Entries[0].Message != request.Message[0] {
		t.Errorf("Unexpected message: %s", response.Entries[0].Message)
	}

	if response.Entries[1].Message != request.Message[1] {
		t.Errorf("Unexpected message: %s", response.Entries[1].Message)
	}
}

func TestStatusByMessageIds(t *testing.T) {
	sms := NewSMSClient(apiKey)

	messageIds := []string{"1234567890", "9876543210"}

	response, err := sms.StatusByMessageIds(messageIds)
	if err != nil {
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}

	if len(response.Entries) != 2 {
		t.Errorf("Unexpected number of entries: %d", len(response.Entries))
	}

	if strconv.FormatInt(response.Entries[0].MessageID, 10) != messageIds[0] {
		t.Errorf("Unexpected message ID: %d", response.Entries[0].MessageID)
	}

	if strconv.FormatInt(response.Entries[1].MessageID, 10) != messageIds[1] {
		t.Errorf("Unexpected message ID: %d", response.Entries[1].MessageID)
	}
}

func TestSelectOutbox(t *testing.T) {
	sms := NewSMSClient(apiKey)

	start := time.Now().Add(-20 * time.Hour)
	end := time.Now()

	response, err := sms.SelectOutbox(start, end)
	if err != nil {
		if err.Error() == "دسترسی به اطلاعات مورد نظر برای شما امکان پذیر نیست" {
			return
		}
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}

	for _, entry := range response.Entries {
		if entry.Date < start.Unix() || entry.Date > end.Unix() {
			t.Errorf("Unexpected date: %d", entry.Date)
		}
	}
}

func TestSelect(t *testing.T) {
	sms := NewSMSClient(apiKey)

	response, err := sms.Select([]string{messageID})
	if err != nil {
		if err.Error() == "دسترسی به اطلاعات مورد نظر برای شما امکان پذیر نیست" {
			return
		}
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}
}

func TestCountOutbox(t *testing.T) {
	sms := NewSMSClient(apiKey)

	start := time.Now().Add(-20 * time.Hour)
	end := time.Now()

	response, err := sms.CountOutbox(start, end)
	if err != nil {
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}
}

func TestLatestOutBox(t *testing.T) {
	sms := NewSMSClient(apiKey)

	response, err := sms.LatestOutBox("", 20)
	if err != nil {
		if err.Error() == "دسترسی به اطلاعات مورد نظر برای شما امکان پذیر نیست" {
			return
		}
		t.Error(err)
	}

	if response.Return.Status != 200 {
		t.Errorf("Unexpected status: %d", response.Return.Status)
	}
}

func TestAccountInfo(t *testing.T) {
	sms := NewSMSClient(apiKey)

	response, err := sms.AccountInfo()
	if err != nil {
		t.Error(err)
	}

	log.Println(response.Return)
}

func TestReceive(t *testing.T) {
	sms := NewSMSClient(apiKey)

	response, err := sms.Receive("10007119", false)
	if err != nil {
		if err.Error() == "ارسال کننده نامعتبر است" {
			return
		}
		t.Error(err)
	}

	log.Println(response.Return)
}
