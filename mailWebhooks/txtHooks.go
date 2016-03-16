package mailWebhooks

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"pingedchat/pcDatabase"
	"strings"
)

// handle post event
func MandrillTextingPost(w http.ResponseWriter, req *http.Request) {
	TRACE.Println("in MandrillTextingPost")
	w.Write([]byte("OK")) // send 200 response

	// decode JSON
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ERROR.Printf("error in ReadAll: ", err)
	}
	strBody := string(body)
	TRACE.Println("body: ", strBody)
	if len(strBody) <= 16 {
		return // can't do much
	}

	unescaped, unescapedErr := url.QueryUnescape(strBody[16:])
	if unescapedErr != nil {
		ERROR.Printf("error in QueryUnescape", unescapedErr)
	}
	// TRACE.Println(string(unescaped))
	var messages MandrillWebHook

	err = json.Unmarshal([]byte(unescaped), &messages)
	if err != nil {
		ERROR.Printf("json.Unmarshal error: ", err)
	}
	TRACE.Println(messages)

	// get db
	db := pcDatabase.Database{}
	db.Connect(nil)
	// loop through each message
	for _, msg := range messages {
		// get user info
		phonenum := msg.Msg.FromEmail[:strings.Index(msg.Msg.FromEmail, "@")]
		TRACE.Println("phonenum in main.go: " + phonenum)
		TRACE.Println("msg.Msg.FromEmail: " + string(msg.Msg.FromEmail))
		user := db.GetAerospikeUserByPhone(string(msg.Msg.FromEmail))
		username := user.Username
		CID := msg.Msg.Email[:strings.Index(msg.Msg.Email, "@")]
		convoMembersStrings := db.GetAerospikeConvoMembers(CID)
		convoMembers := pcDatabase.ToConvoMemberArray(convoMembersStrings)
		var membersArr []string
		for _, e := range convoMembers {
			membersArr = append(membersArr, e.Username)
		}
		membersArrString, err := json.Marshal(membersArr)
		if err != nil {
			ERROR.Printf("error in json.Marshal(membersArr)", err)
		}
		// send message
		data := `{ "cmd" : "SendMessage", "CID" : "` + CID + `", "f_username" : "` + username + `", "t_UIDs" : ` + string(membersArrString) + `, "content" : "` + msg.Msg.Text + `", "m_time": "` + pcDatabase.NowISO8601() + `"}`
		TRACE.Println("data for SendMessage: " + data)
		db.SendMessage(data)
	}

}
