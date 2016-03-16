package pcDatabase

import (
	"database/sql"
	"encoding/json"
	aerospike "github.com/aerospike/aerospike-client-go"
	"github.com/apcera/nats"
	_ "github.com/lib/pq"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"log"
	"os"
	"time"
)

var (
	// for logging
	TRACE = log.New(os.Stdout, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	ERROR = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type Database struct {
	aerospike_conn   *aerospike.Client
	postgres_conn    *sql.DB
	nats_conn        *nats.Conn
	nats_encodedconn *nats.EncodedConn
	// sockjs
	sockjsSession *sockjs.Session
	// channels
	Receive      chan string // public for conn.go
	nats_receive chan string
}

func (db *Database) Connect(sockSession *sockjs.Session) {
	// TRACE.Println("in Database.connect()")
	db.sockjsSession = sockSession
	var err error
	// aerospike
	db.aerospike_conn, err = aerospike.NewClient("127.0.0.1", 3000)
	if err != nil {
		// ERROR.Println("error connecting to aerospike")
		ERROR.Println(err)
	}
	// postgres
	db.postgres_conn, err = sql.Open("postgres", "user=postgres password=postgres dbname=pingedchatdb host=127.0.0.1")
	if err != nil {
		// ERROR.Println("error opening postgres")
		ERROR.Println(err)
	}
	// doesn't open a connection.  Ping it to open a connection.
	// err = db.postgres_conn.Ping()
	// if err != nil {
	// 	ERROR.Println("error connecting to postgres")
	// 	ERROR.Println(err)
	// }
	// TRACE.Println("connecting NATS")
	db.nats_conn, err = nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		// ERROR.Println("error connecting to nats_conn")
		ERROR.Println(err)
	}
	db.nats_encodedconn, err = nats.NewEncodedConn(db.nats_conn, "default")
	if err != nil {
		ERROR.Println("error connecting to nats_encodedconn")
		ERROR.Println(err)
	}

	// make channel for receiving messages used in run()
	db.Receive = make(chan string)
	db.nats_receive = make(chan string)
}

func (db *Database) IsConnected() bool {
	if db.aerospike_conn == nil ||
		db.postgres_conn == nil ||
		db.nats_conn == nil ||
		db.nats_encodedconn == nil {
		return false
	} else {
		return true
	}
}

func (db *Database) IsAerospikePostgresConnected() bool {
	if db.aerospike_conn == nil ||
		db.postgres_conn == nil {
		return false
	} else {
		return true
	}
}

func (db *Database) Close() {
	TRACE.Println("in Database.close()")
	if db.aerospike_conn != nil {
		db.aerospike_conn.Close()
		db.aerospike_conn = nil
	}
	if db.postgres_conn != nil {
		db.postgres_conn.Close()
		db.postgres_conn = nil
	}
	if db.nats_conn != nil {
		db.nats_conn.Close()
		db.nats_conn = nil
	}
	if db.nats_encodedconn != nil {
		db.nats_encodedconn.Close()
		db.nats_encodedconn = nil
	}

	// close channels, will close run() also
	close(db.Receive)
	close(db.nats_receive)
}

func (db *Database) Run() {
	// TRACE.Println("in Database.run()")
	// map of functions in db
	functs := map[string]func(string) string{
		"DeleteUser":                 db.DeleteUser,
		"CreateConversation":         db.CreateConversation,
		"GetConvoData":               db.GetConvoData,
		"GetMoreConvoMessages":       db.GetMoreConvoMessages,
		"GetAllConvoData":            db.GetAllConvoData,
		"GetAllEmails":               db.GetAllEmails,
		"GetUserByUsername":          db.GetUserByUsername,
		"GetUserByEmail":             db.GetUserByEmail,
		"MatchUsers":                 db.MatchUsers,
		"GetPasswordResetUser":       db.GetPasswordResetUser,
		"ChangeUserPassword":         db.ChangeUserPassword,
		"AddToQuota":                 db.AddToQuota,
		"AddToQuotaUsed":             db.AddToQuotaUsed,
		"AddScheduledMessage":        db.AddScheduledMessage,
		"RemoveScheduledMessage":     db.RemoveScheduledMessage,
		"RemoveAllScheduledMessages": db.RemoveAllScheduledMessages,
		"AddFriend":                  db.AddFriend,
		"AcceptFriendRequest":        db.AcceptFriendRequest,
		"DenyFriendRequest":          db.DenyFriendRequest,
		"RemoveFriend":               db.RemoveFriend,
		"SaveAutoreplyMessage":       db.SaveAutoreplyMessage,
		"ChangeProfilePic":           db.ChangeProfilePic,
		"ChangeUserPhone":            db.ChangeUserPhone,
		"ChangeUserEmail":            db.ChangeUserEmail,
		"AddAndroidDev":              db.AddAndroidDev,
		"ChangeAndroidDev":           db.ChangeAndroidDev,
		"RemoveAndroidDev":           db.RemoveAndroidDev,
		"AddIosDev":                  db.AddIosDev,
		"ChangeIosDev":               db.ChangeIosDev,
		"RemoveIosDev":               db.RemoveIosDev,
		"AddFireosDev":               db.AddFireosDev,
		"ChangeFireosDev":            db.ChangeFireosDev,
		"RemoveFireosDev":            db.RemoveFireosDev,
		"RemoveWebDev":               db.RemoveWebDev,
		"GetS3PolicyData":            db.GetS3PolicyData,
		"AddUsersToConversation":     db.AddUsersToConversation,
		"RemoveUserFromConversation": db.RemoveUserFromConversation,
		"ChangeConvoName":            db.ChangeConvoName,
		"UpdateConvoFiles":           db.UpdateConvoFiles,
		"SendMessage":                db.SendMessage,
		"UpdateUserStatus":           db.UpdateUserStatus,
		"UpdateConvoMtime":           db.UpdateConvoMtime,
		"UpdateUnreadCount":          db.UpdateUnreadCount,
		"SendEmailInvite":            db.SendEmailInvite,
		"SendEmailMessage":           db.SendEmailMessage,
		"MarkEmailUnread":            db.MarkEmailUnread,
		"MarkEmailStarred":           db.MarkEmailStarred,
		"AddNewDraft":                db.AddNewDraft,
		"MarkEmailDeleted":           db.MarkEmailDeleted,
		"RemoveDeletedEmails":        db.RemoveDeletedEmails,
	}

	for {
		select {
		case msg, ok := <-db.Receive:
			if !ok {
				ERROR.Println("closing db.run()")
				return
			}
			// TRACE.Println("in msg := <-db.Receive")
			// TRACE.Println(msg)
			cmdJSON := CommonCmdStruct{}
			err := json.Unmarshal([]byte(msg), &cmdJSON)
			if err != nil {
				ERROR.Println(err)
			}
			retstr := `{"retcmd":"bad_command"}`
			if _, ok := functs[cmdJSON.Cmd]; ok {
				// startTime := time.Now().UTC()
				retstr = functs[cmdJSON.Cmd](msg)
				// endTime := time.Now().UTC()
				// TRACE.Println("ElapsedTime in seconds: ", endTime.Sub(startTime))
			} else {
				ERROR.Println("Command " + string(cmdJSON.Cmd) + " was not found and will not be executed.")
			}
			TRACE.Println("str returned from command is: " + string(retstr))
			if retstr != "" {
				(*db.sockjsSession).Send(retstr)
			}
		case msg, ok := <-db.nats_receive:
			if !ok {
				ERROR.Println("closing db.run(), error in db.nats_receive")
				return
			}
			// don't do anything, just send to client
			TRACE.Println("nats_receive: " + string(msg))
			(*db.sockjsSession).Send(msg)
		default:
			// TRACE.Println("db.run() default")
			time.Sleep(50 * time.Millisecond)
		}
	}

}
