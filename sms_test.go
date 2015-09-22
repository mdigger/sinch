package sinch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSMSSend(t *testing.T) {
	sms := &SMS{
		Key:    "f58a02cc-3abf-4cc9-8088-de238200a038",
		Secret: "GcQJltmpf0yQHfHAPk8fhg==",
	}
	// id, err := sms.Send("", "+79031744445", "Проверка связи!")
	// if err != nil {
	// 	t.Fatal("SMS Send error:", err)
	// }
	// fmt.Println("SMS ID:", id)
	// time.Sleep(time.Second * 10)
	id := 115713753
	status, err := sms.Status(id)
	if err != nil {
		t.Fatal("SMS get Status error:", err)
	}
	fmt.Println("Status:", status)
}

// POST /v1/sms/+46700000000
// X-Timestamp: 2014-06-04T13:41:58Z
// Content-Type: application/json
// {"message":"Hello world"}
//
// Content-MD5 = Base64 ( MD5 ( UTF8 ( [BODY] ) ) )
//     jANzQ+rgAHyf1MWQFSwvYw==
// StringToSign
//     POST
//     jANzQ+rgAHyf1MWQFSwvYw==
//     application/json
//     x-timestamp:2014-06-04T13:41:58Z
//     /v1/sms/+46700000000
// Signature = Base64 ( HMAC-SHA256 ( Secret, UTF8 ( StringToSign ) ) )
//     qDXMwzfaxCRS849c/2R0hg0nphgdHciTo7OdM6MsdnM=
// HTTP Authorization Header
//     Authorization: Application 5F5C418A0F914BBC8234A9BF5EDDAD97:qDXMwzfaxCRS849c/2R0hg0nphgdHciTo7OdM6MsdnM=
func TestAuthorization(t *testing.T) {
	sms := &SMS{
		Key:    "5F5C418A0F914BBC8234A9BF5EDDAD97",
		Secret: "JViE5vDor0Sw3WllZka15Q==",
	}
	// кодируем SMS в формат JSON
	data, err := json.Marshal(sinchSMS{
		From:    "",
		Message: "Hello world",
	})
	if err != nil {
		t.Fatal(err)
	}
	// формируем запрос
	req, err := http.NewRequest("POST", sinchURL+"+46700000000", bytes.NewReader(data))
	if err != nil {
		return
	}
	if UserAgent != "" {
		req.Header.Set("User-Agent", UserAgent)
	}
	req.Header.Set("X-Timestamp", "2014-06-04T13:41:58Z") //time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("Accept", "application/json")
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}

	signature, err := sms.sign(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Signature:", signature)
	// println(req.Header.Get("Authorization"))
	req.Write(os.Stdout)
	if req.Header.Get("Authorization") != "Application 5F5C418A0F914BBC8234A9BF5EDDAD97:qDXMwzfaxCRS849c/2R0hg0nphgdHciTo7OdM6MsdnM=" {
		t.Fatal("Bad authorization")
	}
}

func TestIncoming(t *testing.T) {
	sms := &SMS{
		Key:    "8efa3870-ec30-4a55-b612-0a9065d4e5f7",
		Secret: "Ai9PHJVc/UKHpPgiqaZgOA==",
	}
	data := `{
    "event": "incomingSms",
    "to": {
        "type": "number",
        "endpoint": "+46700000000"
    },
    "from": {
        "type": "number",
        "endpoint": "+46700000001"
    },
    "message": "Hello world",
    "timestamp": "2014-12-01T12:00:00Z",
    "version": 1
}`
	// формируем запрос
	req, err := http.NewRequest("POST", "http://localhost:8080/incoming", strings.NewReader(data))
	if err != nil {
		return
	}
	var response = make(map[string]interface{})
	if err := sms.request(req, &response); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Response: %#v", response)
}

func TestAuthorization2(t *testing.T) {
	sms := &SMS{
		Key:    "83d21b0b-605a-4381-b52d-2c27f21317e1",
		Secret: "4YiDmX0WZkedmJQWF7MHsQ==",
	}
	// кодируем SMS в формат JSON
	data, err := json.Marshal(sinchSMS{
		From:    "+14152364961",
		Message: "Test message!",
	})
	if err != nil {
		t.Fatal(err)
	}
	// формируем запрос
	req, err := http.NewRequest("POST", sinchURL+"+79031744445", bytes.NewReader(data))
	if err != nil {
		return
	}
	UserAgent = "MXSMS/0.3.4"
	if UserAgent != "" {
		req.Header.Set("User-Agent", UserAgent)
	}
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("Accept", "application/json")
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}

	signature, err := sms.sign(req)
	if err != nil {
		t.Fatal(err)
	}
	// println(req.Header.Get("Authorization"))
	fmt.Println("Signature:", signature)
	req.Write(os.Stdout)
	// if req.Header.Get("Authorization") != "Application 5F5C418A0F914BBC8234A9BF5EDDAD97:qDXMwzfaxCRS849c/2R0hg0nphgdHciTo7OdM6MsdnM=" {
	// 	t.Fatal("Bad authorization")
	// }
}
