package pcDatabase

import (
	"encoding/json"
	"github.com/lib/pq"
	"time"
)

const (
	POSTGRES_SCHEDULED_MESSAGES_TABLE = "ScheduledMessagesTable"
)

func StartMessagesTicker() {
	// get db
	db := Database{}
	db.Connect(nil)
	// create postgres scheduled messages table, just in case it doesn't exist
	CurString := "CREATE TABLE IF NOT EXISTS " + POSTGRES_SCHEDULED_MESSAGES_TABLE + " (CID varchar NOT NULL, f_username varchar NOT NULL, content varchar NOT NULL, m_time timestamptz NOT NULL, PRIMARY KEY (f_username, m_time) );"
	// TRACE.Println("CurString: " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil when trying to create ScheduledMessagesTable")
	} else {
		_, createerr := db.postgres_conn.Exec(CurString)
		if createerr != nil {
			ERROR.Println("error creating ScheduledMessagesTable: ", createerr)
		}
	}
	// start ticker
	ticker := time.NewTicker(time.Minute * 2)
	go func() {
		for t := range ticker.C {
			TRACE.Println("Tick at ", t)
			go handleScheduledMessages(t, db)
		}
	}()
}

func handleScheduledMessages(t time.Time, db Database) {
	CurString := `DELETE FROM ` + POSTGRES_SCHEDULED_MESSAGES_TABLE + ` WHERE m_time < $1 RETURNING *;`
	TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("in handleScheduledMessages, db.postgres_conn == nil , so returning :(")
		return // can't do anything with no database connection :(
	}
	rows, deleteErr := db.postgres_conn.Query(CurString, t)
	if deleteErr, ok := deleteErr.(*pq.Error); ok {
		ERROR.Println("pq error:", deleteErr.Code.Name())
		ERROR.Println(deleteErr)
	} else {
		defer rows.Close()
		for rows.Next() {
			var SQLCID string
			var SQLF_username string
			var SQLContent string
			var SQLM_time time.Time
			err := rows.Scan(&SQLCID, &SQLF_username, &SQLContent, &SQLM_time)
			if err != nil {
				ERROR.Println(err)
				continue
			} else {
				TRACE.Println("returned from Postgres:  SQLCID: " + string(SQLCID) + " SQLF_username: " + string(SQLF_username) + " SQLContent: " + string(SQLContent) + " SQLM_time: " + SQLM_time.Format(time.RFC3339Nano))
			}
			// need recipients to send the message to
			convoMembersStrings := db.GetAerospikeConvoMembers(SQLCID)
			convoMembers := ToConvoMemberArray(convoMembersStrings)
			var t_UIDs []string
			for _, e := range convoMembers {
				t_UIDs = append(t_UIDs, e.Username)
			}
			// we have all the data we need to send a message
			msg := MessageStruct{
				Cmd:          "SendMessage",
				CID:          SQLCID,
				FromUsername: SQLF_username,
				ToUIDs:       t_UIDs,
				M_time:       SQLM_time.Format(time.RFC3339Nano),
				Content:      SQLContent,
			}
			msgString, msgErr := json.Marshal(msg)
			if msgErr != nil {
				ERROR.Printf("Error in json.Marshal(msg): ", msgErr)
			}
			db.SendMessage(string(msgString))
			// remove from user's scheduled messages
			smcs := ScheduledMessagesCmdStruct{
				Username: SQLF_username,
				CID:      SQLCID,
				Time:     SQLM_time.Format(time.RFC3339Nano),
				Content:  SQLContent,
			}
			smcsString, smcsErr := json.Marshal(smcs)
			if smcsErr != nil {
				ERROR.Printf("Error in json.Marshal(smcs): ", smcsErr)
			}
			db.RemoveScheduledMessage(string(smcsString))
		}
	}
}
