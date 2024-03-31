package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/sms/dto"
)

type SMSClient struct {
	apiKey  string
	baseUrl string
}

const (
	SendSimple = "sms/send.json"
)

func NewSMSClient(apiKey string) *SMSClient {
	return &SMSClient{
		apiKey:  apiKey,
		baseUrl: "https://api.kavenegar.com/v1",
	}
}

func (c *SMSClient) SendSMS(request dto.SendSMSRequest) (response dto.SMSResponse, err error) {
	params := url.Values{}
	params.Add("receptor", strings.Join(request.Receptor, ","))
	params.Add("sender", strings.Join(request.Sender, ","))
	params.Add("message", request.Message)
	baseURL := fmt.Sprintf("%s/%s/sms/send.json", c.baseUrl, c.apiKey)

	reqURL := baseURL + "?" + params.Encode()
	// Build request URL
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) Lookup(request dto.LookupSMSRequest) (response dto.SMSResponse, err error) {
	baseURL := fmt.Sprintf("%s/%s/verify/lookup.json", c.baseUrl, c.apiKey)

	params := url.Values{}
	params.Add("receptor", strings.Join(request.Receptor, ","))
	params.Add("sender", strings.Join(request.Sender, ","))
	params.Add("token", request.Message)
	params.Add("template", request.Template)
	reqURL := baseURL + "?" + params.Encode()
	// Build request URL
	data, err := sendRequest("POST", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) SendBatchSMS(request dto.SendBatchSMSRequest) (response dto.SMSResponse, err error) {
	// Marshal to JSON
	reqBytes, _ := json.Marshal(request)
	baseURL := fmt.Sprintf("%s/%s/sms/sendarray.json", c.baseUrl, c.apiKey)

	data, err := sendRequest("POST", baseURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) StatusByMessageIds(messageIds []string) (response dto.SMSResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/status.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	params.Add("messageid", strings.Join(messageIds, ","))
	reqURL := baseURL + "?" + params.Encode()

	// Build request
	// Make request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) Select(messageIds []string) (response dto.SMSResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/select.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	params.Add("messageid", strings.Join(messageIds, ","))
	reqURL := baseURL + "?" + params.Encode()

	// Build request
	// Make request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) SelectOutbox(start, end time.Time) (response dto.SMSResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/selectoutbox.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	params.Add("startdate", fmt.Sprint(start.Unix()))
	params.Add("enddate", fmt.Sprint(end.Unix()))

	reqURL := baseURL + "?" + params.Encode()

	// Build request
	// Make request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) LatestOutBox(sender string, pageSize int) (response dto.SMSResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/latestoutbox.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	if sender != "" {
		params.Add("sender", sender)
	}

	params.Add("pagesize", strconv.Itoa(pageSize))
	reqURL := baseURL + "?" + params.Encode()

	// Build request
	// Make request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) CountOutbox(start, end time.Time) (response dto.SMSCountResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/countoutbox.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	params.Add("startdate", fmt.Sprint(start.Unix()))
	params.Add("enddate", fmt.Sprint(end.Unix()))

	reqURL := baseURL + "?" + params.Encode()
	// Build request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) CancelByMessageIds(messageIds []string) (response dto.SMSCancelResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/cancel.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	params.Add("messageid", strings.Join(messageIds, ","))
	reqURL := baseURL + "?" + params.Encode()

	// Build request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) Receive(lineNumber string, isread bool) (response dto.SMSResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/receive.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	params.Add("linenumber", lineNumber)
	params.Add("isread", "0")
	if isread {
		params.Add("isread", "1")
	}

	reqURL := baseURL + "?" + params.Encode()

	// Build request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) SMSCountInbox(start, end time.Time) (response dto.SMSCountResponse, err error) {
	// Marshal to JSON

	baseURL := fmt.Sprintf("%s/%s/sms/countinbox.json", c.baseUrl, c.apiKey)
	params := url.Values{}
	params.Add("startdate", fmt.Sprint(start.Unix()))
	params.Add("enddate", fmt.Sprint(end.Unix()))
	reqURL := baseURL + "?" + params.Encode()

	// Build request
	// Make request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func (c *SMSClient) AccountInfo() (response dto.AccountInfoResponse, err error) {
	// Marshal to JSON

	reqURL := fmt.Sprintf("%s/%s/account/info.json", c.baseUrl, c.apiKey)
	// Make request
	data, err := sendRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	return
}

func sendRequest(method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("the following error occurred while sending message:%s \n", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("send", "error", err)
		return nil, err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response dto.HttpResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	if response.Return.Status != 200 {
		return nil, fmt.Errorf("%s", response.Return.Message)
	}

	return data, nil
}
