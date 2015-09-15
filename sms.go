package sinch

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SMS Messaging API URL
const sinchURL = "https://messagingapi.sinch.com/v1/sms/"

// UserAgent string.
var UserAgent = "SinchLib/1.0"

// The SMS Messaging API allows you to send SMS messages to mobile phones and check their status
// using the Sinch platform. You can also rent SMS-enabled numbers from Sinch to receive inbound
// SMS messages from your users that are sent to the backend of your app.
type SMS struct {
	Key       string            // applicationKey
	Secret    string            // applicationSecret
	OnMessage func(IncomingSMS) // on incoming message
	client    http.Client
}

// Send an SMS message to the supplied number, with the contents defined in the msg.
//
// The “From” field indicates the phone number or alphanumeric string that will be displayed to the
// recipient of the SMS message.
//
// You will only be able to send SMS to your verified phone number as long as you have a Sandbox
// app. To send SMS to any phone number, you will need a Production app.
func (s *SMS) Send(from, to, msg string) (msgID int, err error) {
	data, err := json.Marshal(sinchSMS{
		From:    from,
		Message: msg,
	})
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", sinchURL+to, bytes.NewReader(data))
	if err != nil {
		return
	}
	var response = new(sinchResponse)
	if err = s.request(req, response); err != nil {
		return
	}
	msgID = response.MessageID
	return
}

// Status checks the status of a SMS message.
func (s *SMS) Status(msgID int) (status string, err error) {
	req, err := http.NewRequest("GET", sinchURL+strconv.Itoa(msgID), nil)
	if err != nil {
		return
	}
	var response = new(sinchStatus)
	if err = s.request(req, response); err != nil {
		return
	}
	status = response.Status
	return
}

func (s *SMS) request(req *http.Request, response interface{}) error {
	if UserAgent != "" {
		req.Header.Set("User-Agent", UserAgent)
	}
	// The client must send a custom header x-timestamp (time) with each request that is validated
	// by the server. This custom header is used to determine that the request is not too old.
	// The timestamp is also part of the signature. The timestamp must be formatted to ISO 8061
	// specifications.
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("Accept", "application/json")
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}
	// req.SetBasicAuth("application\\"+s.Key, s.Secret) // simple authorization method
	if err := s.sign(req); err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		return json.NewDecoder(resp.Body).Decode(response)
	}
	var errResponse = new(sinchError)
	if err = json.NewDecoder(resp.Body).Decode(errResponse); err != nil {
		return err
	}
	return errResponse
}

func (s *SMS) sign(req *http.Request) error {
	var body string
	if req.Body != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		h := md5.New()
		if _, err := h.Write(data); err != nil {
			return err
		}
		body = base64.StdEncoding.EncodeToString(h.Sum(nil))
	}
	secret, err := base64.StdEncoding.DecodeString(s.Secret)
	if err != nil {
		return err
	}
	sign := strings.Join([]string{
		req.Method,
		body,
		req.Header.Get("Content-Type"),
		"x-timestamp:" + req.Header.Get("X-Timestamp"),
		req.URL.Path,
	}, "\n")
	// log.Print("Sign:\n", sign)
	mac := hmac.New(sha256.New, secret)
	if _, err := io.WriteString(mac, sign); err != nil {
		return err
	}
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	// log.Println("Signature:", signature)
	req.Header.Set("Authorization", fmt.Sprintf("Application %s:%s", s.Key, signature))
	return nil
}

type sinchSMS struct {
	From    string `json:",omitempty"`
	Message string `json:"message"`
}

type sinchResponse struct {
	MessageID int
}

type sinchStatus struct {
	Status string
}

type sinchError struct {
	Code      int `json:"errorCode"`
	Message   string
	Reference string
}

func (e sinchError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// ServeHTTP support incoming SMS event callback.
//
// When a MO SMS is received by the Sinch platform from a specific SMS-enabled number, the system
// sends a notification through a callback request to your backend application. The callback is
// a post request to a specified URL. URLs for callbacks need to be configured in the Sinch portal
// when creating or configuring an application.
func (s *SMS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.Header().Set("Allowed", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var sms = new(IncomingSMS)
	if err := json.NewDecoder(req.Body).Decode(sms); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Body.Close()
	if sms.Event != "incomingSms" {
		http.Error(w, "Not 'incomingSms' event type", http.StatusBadRequest)
		return
	}
	if s.OnMessage != nil {
		s.OnMessage(*sms)
	}
	w.WriteHeader(http.StatusNoContent)
}

// IncomingSMS describe Incoming SMS
type IncomingSMS struct {
	Event     string
	To        Identity
	From      Identity
	Message   string
	Timestamp time.Time
	Version   int
}

type Identity struct {
	Type     string
	Endpoint string
}