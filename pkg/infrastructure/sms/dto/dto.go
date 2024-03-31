package dto

type SendSMSRequest struct {
	Receptor []string `json:"receptor"`
	Sender   []string `json:"sender"`
	Message  string   `json:"message"`
}
type LookupSMSRequest struct {
	Receptor []string `json:"receptor"`
	Sender   []string `json:"sender"`
	Message  string   `json:"message"`
	Template string   `json:"template"`
}

type SendBatchSMSRequest struct {
	Receptor []string `json:"receptor"`
	Sender   []string `json:"sender"`
	Message  []string `json:"message"`
}

// SMSResponse represents send sms response
type SMSResponse struct {
	Return  BaseResponse      `json:"return"`
	Entries []SMSResponseItem `json:"entries"`
}

type HttpResponse struct {
	Return BaseResponse `json:"return"`
}

type BaseResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type SMSResponseItem struct {
	MessageID  int64   `json:"messageid"`
	Message    string  `json:"message"`
	Status     int     `json:"status"`
	StatusText string  `json:"statustext"`
	Sender     string  `json:"sender"`
	Receptor   string  `json:"receptor"`
	Date       int64   `json:"date"`
	Cost       float32 `json:"cost"`
}

// Define response struct
type SMSCountResponse struct {
	Return  BaseResponse
	Entries []SMSCountInfo
}

type SMSCountInfo struct {
	StartDate int64
	EndDate   int64
	SumPart   int64
	SumCount  int64
	Cost      int64
}

// Define Response struct
type SMSCancelResponse struct {
	Return  BaseResponse
	Entries []SMSCancelResult
}
type SMSCancelResult struct {
	MessageID  int64
	Status     int
	StatusText string
}

// Define response struct
type AccountInfoResponse struct {
	Return  BaseResponse
	Entries struct {
		RemainCredit int
		ExpireDate   int64
		Type         string
	}
}
