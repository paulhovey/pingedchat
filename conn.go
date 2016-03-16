package main

import (
	"encoding/json"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"pingedchat/pcDatabase"
)

func sockHandler(session sockjs.Session) {
	TRACE.Println("new sockjs session established")

	db := pcDatabase.Database{}
	db.Connect(&session)
	defer db.Close()

	var validatedUser bool = false
	var token string
	ValidUser := pcDatabase.UserStruct{}
	cmd := pcDatabase.CreateUserCmdStruct{}
	for !validatedUser {
		// TRACE.Println("waiting for valid user")
		if msg, err := session.Recv(); err == nil {
			TRACE.Println("message received: " + msg)
			err := json.Unmarshal([]byte(msg), &cmd)
			if err != nil {
				ERROR.Println(err)
			}
			TRACE.Println("cmd.Cmd = " + cmd.Cmd)
			if cmd.Cmd == "CreateUser" {
				ValidUser = db.CreateUser(msg)
				if ValidUser.UsernameUpper == "" {
					// invalid user
					session.Send(`{"cmd":"CreateUser", "error":"Username already taken."}`)
				} else {
					validatedUser = true
					token = cmd.Token
					session.Send(ValidUser.ToJSONStringWithCmd("CreateUser"))
				}
			} else if cmd.Cmd == "ValidateUser" {
				ValidUser = db.ValidateUser(msg)
				if ValidUser.UsernameUpper != "" {
					validatedUser = true
					token = cmd.Token
				}
				session.Send(ValidUser.ToJSONStringWithCmd("ValidateUser")) // send back no matter what so we can show bad password on UI
			} else if cmd.Cmd == "GetPasswordResetUser" {
				session.Send(db.GetPasswordResetUser(msg))
			} else if cmd.Cmd == "ResetUserPassword" {
				session.Send(db.ResetUserPassword(msg))
			}

		} else {
			return // exit this place
		}
	}
	// TRACE.Println("found valid user")
	// TRACE.Println(ValidUser.ToJSONString())
	// we have a valid user at this point
	// web token is already linked
	// start our db handler
	// go pcDatabase.Protect(db.Run)
	go db.Run()
	defer db.RemoveUserWebToken(ValidUser.UsernameUpper, token)

	for {
		msg, err := session.Recv()
		if err != nil {
			// ERROR.Println("error in for session.Recv() conn.go")
			// ERROR.Println(err)
			break
		}
		TRACE.Println("message received: " + msg)
		db.Receive <- msg
	}

}
