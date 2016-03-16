package pcDatabase

import (
	"github.com/mostafah/mandrill"
	"github.com/timehop/apns"
	"pingedchat/pcDatabase/gcm"
)

const (
	ANDROID_API_KEY     = "FILL_ME"
	MANDRILL_API_KEY    = "FILL_ME"
	TELAPI_TESTING_SID  = "FILL_ME"
	TELAPI_TESTING_AUTH = "FILL_ME"

	// generate .pem files for Apple certificates
	APNS_CERT_FILE = "pcDatabase/iosDevelopmentPushCert.pem"
	APNS_KEY_FILE  = "pcDatabase/iosDevelopmentPushKey-noenc.pem"
)

var (
	admToken   ADMResponse
	APNSClient apns.Client
)

func CreatePushTitle(f_username string, content string) string {
	title := ""
	// some error checking
	if len(content) < 4 {
		title = f_username + " - " + content
	} else {
		if content[:2] == "<a" {
			title = f_username + " sent you a link!"
		} else if content[:4] == "<img" {
			title = f_username + " sent you a picture!"
		} else {
			// can't slice what's not there
			if len(content) < 40 {
				title = f_username + " - " + content // substring 40 characters limit
			} else {
				title = f_username + " - " + content[:40] // substring 40 characters limit
			}
		}
	}
	return title
}

func PushToAndroid(droids []string, messageData MessageStruct) {
	TRACE.Println("droids:")
	TRACE.Println(droids)
	// taken from https://github.com/alexjlockwood/gcm
	// create message title
	title := CreatePushTitle(messageData.FromUsername, messageData.Content)
	data := map[string]interface{}{
		"message": messageData.Content,
		// "content":    messageData.Content,
		"msgcnt":     1, // apparently https://github.com/phonegap-build/PushPlugin wants this
		"title":      messageData.FromUsername,
		"CID":        messageData.CID,
		"f_username": messageData.FromUsername,
		"m_time":     messageData.M_time,
		"cmd":        messageData.Cmd,
	}
	msg := gcm.NewMessage(data, title, "4", droids...)
	sender := &gcm.Sender{ApiKey: ANDROID_API_KEY}
	// Send the message and receive the response after at most two retries.
	response, err := sender.Send(msg, 2)
	if err != nil {
		ERROR.Println("Failed to send gcm message:", err)
		return
	} else {
		TRACE.Printf("response from pushing to android: success: %d  failure: %d\n", response.Success, response.Failure)
	}
}

func PushToFireos(fires []string, messageData MessageStruct) {
	for _, fire := range fires {
		TRACE.Println("fire device " + fire)
		title := CreatePushTitle(messageData.FromUsername, messageData.Content)
		msg := new(ADMMessageStruct)
		msg.Data = make(map[string]string, 6) // we have 6 things to send
		// "message" is what's displayed in the notification bar
		msg.Data["message"] = title
		msg.Data["content"] = messageData.Content
		msg.Data["m_time"] = messageData.M_time
		msg.Data["CID"] = messageData.CID
		msg.Data["f_username"] = messageData.FromUsername
		// msg.Data["title"] = messageData.FromUsername
		msg.Data["cmd"] = messageData.Cmd
		msg.MsgGroup = "PingedChat"
		msg.TTL = 3600

		TRACE.Println("messageData.Content: " + messageData.Content)
		TRACE.Println("messageData.M_time: " + messageData.M_time)
		TRACE.Println("messageData.CID: " + messageData.CID)
		TRACE.Println("messageData.FromUsername: " + messageData.FromUsername)
		TRACE.Println(msg.Data)
		sendADM(fire, msg)
	}

}

func PushToIos(IosDevices []string, messageData MessageStruct) {
	// taken from https://github.com/anachronistic/apns
	if APNSClient.Conn == nil {
		var err error
		// APNSClient = apns.NewClient("gateway.sandbox.push.apple.com:2195", "YOUR_CERT_PEM", "YOUR_KEY_NOENC_PEM") // TODO
		// how-to for certs at https://github.com/joekarl/go-libapns
		// basically:
		// Separate pem files from p12
		// openssl pkcs12 -clcerts -nokeys -out cert.pem -in cert.p12
		// openssl pkcs12 -nocerts -out key.pem -in key.p12
		// Remove password from pem file
		// openssl rsa -in key.pem -out key-noenc.pem
		APNSClient, err = apns.NewClientWithFiles(apns.SandboxGateway, APNS_CERT_FILE, APNS_KEY_FILE)
		if err != nil {
			ERROR.Printf("Could not create apns client", err.Error())
		} else {
			TRACE.Println("Created new APNS client successfully")
		}

		// start feedback loop
		go func() {
			f, err := apns.NewFeedbackWithFiles(apns.SandboxFeedbackGateway, APNS_CERT_FILE, APNS_KEY_FILE)
			if err != nil {
				ERROR.Printf("Could not create feedback", err.Error())
			} else {
				TRACE.Println("APNS feedback created successfully")
			}

			for ft := range f.Receive() {
				TRACE.Println("APNS Feedback for token:", ft.DeviceToken)
			}
		}()

		// start error handling
		go func() {
			for f := range APNSClient.FailedNotifs {
				ERROR.Println("APNS Notif", f.Notif.ID, "failed with", f.Err.Error())
			}
		}()
	}

	// data := map[string]interface{}{
	// 	"message":    messageData.Content,
	// 	"msgcnt":     1, // apparently https://github.com/phonegap-build/PushPlugin wants this
	// 	"title":      messageData.FromUsername,
	// 	"CID":        messageData.CID,
	// 	"f_username": messageData.FromUsername,
	// 	"m_time":     messageData.M_time,
	// 	"cmd":        "send_message",
	// }

	title := CreatePushTitle(messageData.FromUsername, messageData.Content)

	for _, iOSDev := range IosDevices {
		TRACE.Println("iOSDev = " + iOSDev)

		p := apns.NewPayload()
		p.APS.Alert.Body = title
		p.APS.Badge.Set(1)
		p.APS.ContentAvailable = 1
		p.APS.Sound = "turn_down_for_what.aiff"

		p.SetCustomValue("content", messageData.Content)
		p.SetCustomValue("CID", messageData.CID)
		p.SetCustomValue("f_username", messageData.FromUsername)
		p.SetCustomValue("m_time", messageData.M_time)
		p.SetCustomValue("cmd", messageData.Cmd)

		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = iOSDev
		m.Priority = apns.PriorityImmediate

		APNSClient.Send(m)
	}

}

func PushToSMS(phonenum string, messageData MessageStruct, convoName string, toUsername string) {
	mandrill.Key = MANDRILL_API_KEY
	// you can test your API key with Ping
	err := mandrill.Ping()
	// everything is OK if err is nil
	if err != nil {
		ERROR.Println("error in PushToSMS mandrill.Ping()")
		ERROR.Println(err)
	}
	TRACE.Println("PushToSMS, phonenum = " + phonenum)
	msg := mandrill.NewMessageTo(phonenum, toUsername)
	msg.HTML = messageData.Content
	//msg.Text = messageData.Content
	msg.Subject = convoName
	msg.FromEmail = messageData.CID + "@txt.pingedchat.com"
	//msg.FromName = messageData.CID + "@txt.pingedchat.com"
	res, err := msg.Send(false)
	if err != nil {
		ERROR.Println("error in PushToSMS msg.Send(false)")
		ERROR.Println(err)
		if res[0].Status != "sent" {
			ERROR.Println("res.Status in PushToSMS = " + res[0].Status)
			ERROR.Println("res.RejectionReason in PushToSMS = " + res[0].RejectionReason)
		}
	}
}

// helper functions
// run periodically updates the adm access token.
/*
func UpdateADM() {
    go updateToken()

    ticker := time.NewTicker(time.Second * 1800)
    for {
        select {
        case <-ticker.C:
            attempts := 0
            for {
                if attempts > 5 {
                    // give up and try at next tick
                    break
                }
                err := updateToken()
                if err != nil {
                    ERROR.Println(err)
                    time.Sleep(time.Second)
                    attempts++
                    continue
                }
                break
            }
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

    request, err := http.NewRequest("POST", "https://api.amazon.com/auth/O2/token", bytes.NewBufferString(data.Encode()))
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

    r := ADMResponse{}
    err = json.Unmarshal(body, &r)
    if err != nil {
        ERROR.Println(err, string(body))
        return err
    }
    TRACE.Printf("adm %+v", r)
    admToken = r
    return nil
}
*/
