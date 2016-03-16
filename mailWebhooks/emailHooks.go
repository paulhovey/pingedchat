package mailWebhooks

import (
	"encoding/base64"
	aerospike "github.com/aerospike/aerospike-client-go"
	_ "github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"pingedchat/pcDatabase"
	"strconv"
	"strings"
	"time"
)

var (
	// for logging
	TRACE = log.New(os.Stdout, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	ERROR = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

var db pcDatabase.Database

// handle post event
func MandrillEmailPost(w http.ResponseWriter, req *http.Request) {
	TRACE.Println("in MandrillEmailPost")
	w.Write([]byte("OK")) // send 200 response

	// decode JSON
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ERROR.Printf("error in ReadAll: ", err)
	}
	strBody := string(body)
	if len(strBody) <= 16 {
		TRACE.Println("body was less than 16 characters: ", strBody)
		return // can't do much
	}

	unescaped, unescapedErr := url.QueryUnescape(strBody[16:])
	if unescapedErr != nil {
		ERROR.Printf("error in QueryUnescape", unescapedErr)
	}
	// TRACE.Println("unescaped body: ", unescaped)
	var messages MandrillWebHook

	err = ffjson.Unmarshal([]byte(unescaped), &messages)
	if err != nil {
		ERROR.Printf("ffjson.Unmarshal error: ", err)
	}
	// TRACE.Println(messages)

	// get db
	if !db.IsAerospikePostgresConnected() {
		db.Connect(nil)
	}

	// loop through each message
	for _, msg := range messages {
		// get user info
		username := msg.Msg.Email[:strings.Index(msg.Msg.Email, "@pinged.email")]
		user := db.GetAerospikeUser(username)
		TRACE.Println("username:" + username)
		TRACE.Println("username found: " + user.Username)
		if user.Username == "" {
			continue
		}
		TRACE.Println("Received message:")
		// create ToEmails structure
		var ToEmails []string
		for _, toEmail := range msg.Msg.To {
			TRACE.Println("toEmail: ", toEmail)
			// see http://stackoverflow.com/questions/4473804/is-it-correct-that-sometimes-i-get-angle-brackets-in-from-field-of-an-e-mail-m
			if string(toEmail[1]) == "" {
				ToEmails = append(ToEmails, toEmail[0]) // just save email address
			} else {
				ToEmails = append(ToEmails, toEmail[1]+" <"+toEmail[0]+">") // save name <email>
			}
		}
		ToEmailBytes, err := ffjson.Marshal(ToEmails)
		ToEmailStr := string(ToEmailBytes)
		if err != nil {
			ERROR.Println("error in ffjson.Marshal in MandrillEmailPost() marshalling ToEmails string")
			ERROR.Println(err)
		}

		// print debug info
		TRACE.Println("From:    " + msg.Msg.FromEmail)
		TRACE.Println("To:      " + ToEmailStr)
		TRACE.Println("To:      " + msg.Msg.Email)
		TRACE.Println("Subject: " + msg.Msg.Subject)
		TRACE.Println("Message: " + msg.Msg.Text)
		TRACE.Println("Attachments: ", msg.Msg.Attachments)

		// loop through attachments, upload to s3
		var mail_attachments []string
		for filename, attachment := range msg.Msg.Attachments {
			var sDec []byte
			if attachment.IsBase64 {
				sDec, _ = base64.StdEncoding.DecodeString(attachment.Content)
			} else {
				sDec = []byte(attachment.Content)
			}

			// remove all spaces with underscores
			filename = strings.Replace(filename, " ", "_", -1)

			fullS3FilePath := db.UploadToS3(sDec, filename, user)
			if fullS3FilePath != "" {
				mail_attachments = append(mail_attachments, fullS3FilePath)
			}
		} // end for filename, attachment := range msg.Msg.Attachments

		isSpam := (bool)(msg.Msg.SpamReport.Score >= 5.0)

		// create timestamp string
		t := time.Unix(int64(msg.Ts), 0)
		t_s := t.Format(pcDatabase.ISO8601_SECONDS) // time string

		// create attachments string
		att_bytes, err := ffjson.Marshal(mail_attachments)
		att_str := string(att_bytes)
		if err != nil {
			ERROR.Println("error in ffjson.Marshal in MandrillEmailPost() marshalling attachments string")
			ERROR.Println(err)
			att_str = "[]"
		}
		TRACE.Println("att_str: ", att_str)
		msg_html := strings.TrimSpace(msg.Msg.Html)
		msg_html = strings.Replace(msg_html, "\n", "<br/>", -1)
		// call helper function
		db.AddEmailToDb(user.Username, msg.Msg.FromEmail, ToEmailStr, msg.Msg.Email, msg.Msg.Subject, msg_html, att_str, false, true, isSpam, t_s, t_s)

		// send instant update to user
		ret_str := `{"cmd":"EmailReceived",` +
			`"FromEmail":"` + msg.Msg.FromEmail + `",` +
			`"ToEmails":` + ToEmailStr + `,` +
			`"RecvEmail":"` + msg.Msg.Email + `",` +
			`"Subject":"` + strings.TrimSpace(msg.Msg.Subject) + `",` +
			`"Content":"` + strings.TrimSpace(html.EscapeString(msg_html)) + `",` +
			`"Attachments":` + att_str + `,` +
			`"Spam":"` + strconv.FormatBool(isSpam) + `",` +
			`"RecvTime":"` + t_s + `"` +
			`}`
		db.SendStringToWebDevices(user.Web, ret_str)
		// update m_time for user
		emailMtimeBin := aerospike.NewBin("EmailMtime", t_s)
		key, err := aerospike.NewKey(pcDatabase.AEROSPIKE_USERS_NAMESPACE, pcDatabase.AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, emailMtimeBin)
		}
	} // end for _, msg := range messages
}
