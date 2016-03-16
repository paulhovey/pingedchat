package pcDatabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	ADMServer ADMServerStruct // keep you around!

	AMAZON_ADM_URL   = "https://api.amazon.com/messaging/registrations/"
	AMAZON_TOKEN_URL = "https://api.amazon.com/auth/O2/token"

	ADMDebugClientID      = "FILL_ME"
	ADMProductionClientID = "FILL_ME"

	ADMDebugClientSecret      = "FILL_ME" // MBP
	ADMProductionClientSecret = "FILL_ME"
)

type ADMServerStruct struct {
	Token ADMTokenResponseStruct
}

type ADMTokenResponseStruct struct {
	Scope       string `json:"scope"`        // should be "messaging:push"
	TokenType   string `json:"token_type"`   // should be "bearer"
	ExpiresIn   int    `json:"expires_in"`   // default 3600
	AccessToken string `json:"access_token"` // should be "Atc|xxxxxxx"
}

type ADMMessageStruct struct {
	Data     map[string]string `json:"data"`
	MsgGroup string            `json:"consolidationKey,omitempty"`
	TTL      int64             `json:"expiresAfter,omitempty"`
	MD5      string            `json:"md5,omitempty"`
}

// ADMResponse https://developer.amazon.com/public/apis/engage/device-messaging/tech-docs/06-sending-a-message
type ADMResponseStruct struct {
	StatusCode int   `json:"statusCode"`
	Error      error `json:"error"`
	// The calculated base-64-encoded MD5 checksum of the data field.
	MD5 string `json:"md5"`
	// A value created by ADM that uniquely identifies the request.
	// In the unlikely event that you have problems with ADM,
	// Amazon can use this value to troubleshoot the problem.
	RequestID string `json:"requestID"`
	// This field is returned in the case of a 429, 500, or 503 error response.
	// Retry-After specifies how long the service is expected to be unavailable.
	// This value can be either an integer number of seconds (in decimal) after
	// the time of the response or an HTTP-format date. See the HTTP/1.1
	// specification, section 14.37, for possible formats for this value.
	RetryAfter int `json:"retryAfter"`
	// The current registration ID of the app instance.
	// If this value is different than the one passed in by
	// your server, your server must update its records to use this value.
	RegistrationID string `json:"registrationID"`
	// 400
	//   InvalidRegistrationId
	//   InvalidData
	//   InvalidConsolidationKey
	//   InvalidExpiration
	//   InvalidChecksum
	//   InvalidType
	//   Unregistered
	// 401
	//   AccessTokenExpired
	// 413
	//   MessageTooLarge
	// 429
	//   MaxRateExceeded
	// 500
	//   n/a
	Reason string `json:"reason"`
}

// initialize struct update token
func ADMinit() {
	go updateToken()    // initial call
	go Protect(ADMLoop) // keep it going into the future
}

func ADMLoop() {
	ticker := time.NewTicker(time.Second * 3600) // taken from example at https://developer.amazon.com/public/apis/engage/device-messaging/tech-docs/05-requesting-an-access-token
	for t := range ticker.C {
		TRACE.Println("Tick in ADMinit at ", t)
		attempts := 0
		for {
			if attempts > 5 {
				// give up and try at next tick
				break
			}
			err := updateToken()
			if err != nil {
				ERROR.Printf("error in updateToken: ", err)
				time.Sleep(time.Second)
				attempts++
				continue
			}
			break
		}
	}
}

// updateToken updates the adm access token.
func updateToken() error {
	TRACE.Println("getting new access token")

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "messaging:push")
	data.Set("client_id", ADMDebugClientID)
	data.Set("client_secret", ADMDebugClientSecret)

	request, err := http.NewRequest("POST", AMAZON_TOKEN_URL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		ERROR.Println(err)
		return err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		ERROR.Printf("%s %+v", err, resp)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ERROR.Println(err)
		return err
	}

	r := ADMTokenResponseStruct{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		ERROR.Println(err, string(body))
		return err
	}
	TRACE.Printf("ADM token response: %v", r)
	ADMServer.Token = r
	return nil
}

// now for actually sending the message!
func sendADM(regid string, messageData *ADMMessageStruct) {
	// check token
	if ADMServer.Token.AccessToken == "" {
		updateToken() // get a new token if it doesn't exist
	}
	TRACE.Println("ADMServer.Token.AccessToken: " + ADMServer.Token.AccessToken)

	POSTUrl := fmt.Sprintf("%v%v/messages", AMAZON_ADM_URL, regid)
	TRACE.Println("POSTUrl: " + POSTUrl)

	// need to Marshal our data
	msgdata, err := json.Marshal(messageData)
	if err != nil {
		ERROR.Printf("Error in json.Marshal(messageData): ", err)
		return
	}
	TRACE.Println("\n\nmsgdata: " + string(msgdata) + "\n\n")

	req, err := http.NewRequest("POST", POSTUrl, bytes.NewBuffer(msgdata))
	if err != nil {
		ERROR.Printf("Error in http.NewRequest('POST'): ", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Amzn-Type-Version", "com.amazon.device.messaging.ADMMessage@1.0")
	req.Header.Set("X-Amzn-Accept-Type", "com.amazon.device.messaging.ADMSendResult@1.0")
	req.Header.Set("Authorization", "Bearer "+ADMServer.Token.AccessToken)

	TRACE.Println("req:")
	TRACE.Println(req)

	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ERROR.Printf("Error after client.Do(req): ", err)
		return
	}

	TRACE.Println("resp returned from client.Do(req):")
	TRACE.Println(resp)

	if resp.StatusCode != 200 {
		ERROR.Printf("resp.StatusCode = ", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ERROR.Printf("err after ioutil.ReadAll(resp.Body): ", err)
		return
	}

	TRACE.Println("resp.Body: " + string(content))

}
