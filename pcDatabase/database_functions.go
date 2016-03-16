package pcDatabase

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	aerospike "github.com/aerospike/aerospike-client-go"
	"github.com/lib/pq"
	"github.com/mostafah/mandrill"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
	"html"
	"sort"
	"strings"
	"time"
)

const (
	AEROSPIKE_USERS_NAMESPACE      = "users"
	AEROSPIKE_USERS_USERNAME_TABLE = "username"
	AEROSPIKE_USERS_ACTIVE_TABLE   = "active"
	AEROSPIKE_USERS_PHONES_TABLE   = "phones"

	AEROSPIKE_USERS_USERNAME_EMAIL_BIN = "Email"
	AEROSPIKE_USERS_USERNAME_PHONE_BIN = "Phone"

	AEROSPIKE_CONVOS_NAMESPACE   = "convos"
	AEROSPIKE_CONVOS_MEMBERS_KEY = "members"
	AEROSPIKE_CONVOS_NAME_KEY    = "name"
	AEROSPIKE_CONVOS_MTIME_KEY   = "m_time"
	AEROSPIKE_CONVOS_FILES_KEY   = "files"
)

// Database functions
func (db *Database) ReadAerospike(key *aerospike.Key) *aerospike.Record {
	if db.aerospike_conn != nil {
		rec, err := db.aerospike_conn.Get(nil, key)
		if err == nil {
			return rec
		} else {
			ERROR.Println("error in ReadAerospike")
			ERROR.Println(err)
			return nil
		}
	} else {
		ERROR.Println("db.aerospike_conn == nil in ReadAerospike")
		return nil
	}
}

func (db *Database) WriteAerospikeMultipleBins(key *aerospike.Key, value aerospike.BinMap) bool {
	if db.aerospike_conn != nil {
		writePolicy := &aerospike.WritePolicy{
			BasePolicy:         *aerospike.NewPolicy(),
			RecordExistsAction: aerospike.UPDATE,
			GenerationPolicy:   aerospike.NONE,
			CommitLevel:        aerospike.COMMIT_ALL,
			Generation:         0,
			Expiration:         0,
			SendKey:            false,
		}
		err := db.aerospike_conn.Put(writePolicy, key, value)
		if err == nil {
			return true
		} else {
			ERROR.Println("error in WriteAerospikeMultipleBins")
			ERROR.Println(err)
			return false
		}
	} else {
		ERROR.Println("db.aerospike_conn == nil in WriteAerospikeMultipleBins")
		return false
	}
}

func (db *Database) UpdateAerospikeSingleBin(key *aerospike.Key, value *aerospike.Bin) bool {
	if db.aerospike_conn != nil {
		writePolicy := &aerospike.WritePolicy{
			BasePolicy:         *aerospike.NewPolicy(),
			RecordExistsAction: aerospike.UPDATE, // https://github.com/aerospike/aerospike-client-go/blob/master/record_exists_action.go
			GenerationPolicy:   aerospike.NONE,
			CommitLevel:        aerospike.COMMIT_ALL,
			Generation:         1,
			Expiration:         0,
			SendKey:            false,
		}
		err := db.aerospike_conn.PutBins(writePolicy, key, value)
		if err == nil {
			return true
		} else {
			ERROR.Println("error in UpdateAerospikeSingleBin")
			ERROR.Println(err)
			return false
		}
	} else {
		ERROR.Println("db.aerospike_conn == nil in UpdateAerospikeSingleBin")
		return false
	}
}

func (db *Database) ReplaceAerospikeSingleBin(key *aerospike.Key, value *aerospike.Bin) bool {
	if db.aerospike_conn != nil {
		writePolicy := &aerospike.WritePolicy{
			BasePolicy:         *aerospike.NewPolicy(),
			RecordExistsAction: aerospike.REPLACE, // https://github.com/aerospike/aerospike-client-go/blob/master/record_exists_action.go
			GenerationPolicy:   aerospike.NONE,
			CommitLevel:        aerospike.COMMIT_ALL,
			Generation:         1,
			Expiration:         0,
			SendKey:            false,
		}
		err := db.aerospike_conn.PutBins(writePolicy, key, value)
		if err == nil {
			return true
		} else {
			ERROR.Println("error in ReplaceAerospikeSingleBin")
			ERROR.Println(err)
			return false
		}
	} else {
		ERROR.Println("db.aerospike_conn == nil in ReplaceAerospikeSingleBin")
		return false
	}
}

func (db *Database) DeleteAerospike(key *aerospike.Key) bool {
	if db.aerospike_conn != nil {
		existed, err := db.aerospike_conn.Delete(nil, key)
		if err == nil {
			return existed
		} else {
			ERROR.Println("error in DeleteAerospike")
			ERROR.Println(err)
			return false
		}
	} else {
		ERROR.Println("db.aerospike_conn == nil in DeleteAerospike")
		return false
	}
}

// getters
func (db *Database) GetAerospikeUser(username string) UserStruct {
	defer func() {
		if r := recover(); r != nil {
			ERROR.Println("Recovered in GetAerospikeUser", r)
		}
	}()
	if strings.TrimSpace(username) == "" {
		return UserStruct{} // return empty user if they passed in an empty username
	}
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(username))
	if err == nil {
		rec := db.ReadAerospike(key)
		if rec != nil {
			user := FillUserWithAerospikeBins(rec.Bins)
			// TRACE.Println("In GetAerospikeUser, user = " + user.ToJSONString())
			return user
		} else {
			return UserStruct{}
		}
	} else {
		ERROR.Println(err)
		return UserStruct{}
	}
}

func (db *Database) GetAerospikeUserActive(username string) *aerospike.Record {
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_ACTIVE_TABLE, strings.ToUpper(username))
	if err == nil {
		rec := db.ReadAerospike(key)
		return rec
	} else {
		ERROR.Println(err)
		return nil
	}
}

/*
//   NOW USING AQL TO QUERY INSTEAD OF ANOTHER TABLE
func (db *Database) GetAerospikeUserByPhone(phonenum string) string {
    key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_PHONES_TABLE, phonenum)
    if err == nil {
		rec := db.ReadAerospike(key)
		Username := rec.Bins["Username"].(string)
		return Username
    } else {
        ERROR.Println(err)
        return ""
    }
}
*/

func (db *Database) GetAerospikeUserByPhone(phonenum string) UserStruct {
	defer func() {
		if r := recover(); r != nil {
			ERROR.Println("Recovered in GetAerospikeUserByPhone", r)
		}
	}()
	stmt := aerospike.NewStatement(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE)
	stmt.Addfilter(aerospike.NewEqualFilter(AEROSPIKE_USERS_USERNAME_PHONE_BIN, phonenum))
	if db.aerospike_conn == nil {
		return UserStruct{} // can't do anything without a connection :(
	}
	// now query
	rs, err := db.aerospike_conn.Query(nil, stmt)
	// deal with error
	if err != nil {
		ERROR.Println("Error on db.aerospike_conn.Query in GetAerospikeUserByPhone: ")
		ERROR.Println(err)
	}

	// should only be one result
	for res := range rs.Results() {
		if res.Err != nil {
			// handle error here
			// if you want to exit, cancel the recordset to release the resources
			ERROR.Println("Error on db.aerospike_conn.Query in res.Err: ")
			ERROR.Println(res.Err)
			rs.Close()
			return UserStruct{} // there was an error
		} else {
			// process record here
			// TRACE.Println("res.Record.Bins: ")
			// TRACE.Println(res.Record.Bins)
			user := FillUserWithAerospikeBins(res.Record.Bins)
			rs.Close()
			return user
		}
	}
	// if here, we didn't return a user above, so return nothing
	return UserStruct{}
}

func (db *Database) GetAerospikeUserByEmail(email string) UserStruct {
	defer func() {
		if r := recover(); r != nil {
			ERROR.Println("Recovered in GetAerospikeUserByEmail", r)
		}
	}()
	stmt := aerospike.NewStatement(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE)
	stmt.Addfilter(aerospike.NewEqualFilter(AEROSPIKE_USERS_USERNAME_EMAIL_BIN, email))
	// TRACE.Println("in GetAerospikeUserByEmail, email = " + email)
	if db.aerospike_conn == nil {
		return UserStruct{} // can't do anything without a connection :(
	}
	// now query
	rs, err := db.aerospike_conn.Query(nil, stmt)
	// deal with error
	if err != nil {
		ERROR.Println("Error on db.aerospike_conn.Query in GetAerospikeUserByEmail: ")
		ERROR.Println(err)
	}

	// should only be one result
	for res := range rs.Results() {
		if res.Err != nil {
			// handle error here
			// if you want to exit, cancel the recordset to release the resources
			ERROR.Println("Error on db.aerospike_conn.Query in res.Err: ")
			ERROR.Println(res.Err)
			rs.Close()
			return UserStruct{} // there was an error
		} else {
			// process record here
			// TRACE.Println("res.Record.Bins: ")
			// TRACE.Println(res.Record.Bins)
			user := FillUserWithAerospikeBins(res.Record.Bins)
			rs.Close()
			return user
		}
	}
	// if here, we didn't return a user above, so return nothing
	return UserStruct{}
}

func (db *Database) GetAerospikeConvoMembers(CID string) []string {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_MEMBERS_KEY)
	if err == nil {
		rec := db.ReadAerospike(key)
		if rec != nil {
			members := InterfaceArrayToStringArray(rec.Bins["Members"].([]interface{}))
			return members
		} else {
			return make([]string, 0)
		}
	} else {
		ERROR.Println(err)
		return make([]string, 0)
	}
}

func (db *Database) GetAerospikeConvoName(CID string) string {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_NAME_KEY)
	if err == nil {
		rec := db.ReadAerospike(key)
		if rec != nil {
			name := rec.Bins["Name"].(string)
			return name
		} else {
			return ""
		}
	} else {
		ERROR.Println(err)
		return ""
	}
}

func (db *Database) GetAerospikeConvoMtime(CID string) string {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_MTIME_KEY)
	if err == nil {
		rec := db.ReadAerospike(key)
		if rec != nil {
			m_time := rec.Bins["M_time"].(string)
			return m_time
		} else {
			return ""
		}
	} else {
		ERROR.Println(err)
		return ""
	}
}

func (db *Database) GetAerospikeConvoFiles(CID string) []string {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_FILES_KEY)
	if err == nil {
		rec := db.ReadAerospike(key)
		if rec != nil {
			files := InterfaceArrayToStringArray(rec.Bins["Files"].([]interface{}))
			return files
		} else {
			return make([]string, 0)
		}
	} else {
		ERROR.Println(err)
		return make([]string, 0)
	}
}

// setters
func (db *Database) SetAerospikeUser(username string, data aerospike.BinMap) bool {
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(username))
	if err == nil {
		return db.WriteAerospikeMultipleBins(key, data)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) SetAerospikeUserActive(username string, devices aerospike.BinMap) bool {
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_ACTIVE_TABLE, strings.ToUpper(username))
	if err == nil {
		return db.WriteAerospikeMultipleBins(key, devices)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) SetAerospikeUserPhone(phonenum string, user aerospike.BinMap) bool {
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_PHONES_TABLE, phonenum)
	if err == nil {
		return db.WriteAerospikeMultipleBins(key, user)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) SetAerospikeConvoMembers(CID string, members []string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_MEMBERS_KEY)
	if err == nil {
		bins := aerospike.BinMap{
			"Members": members,
		}
		return db.WriteAerospikeMultipleBins(key, bins)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) SetAerospikeConvoName(CID string, name string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_NAME_KEY)
	if err == nil {
		bins := aerospike.BinMap{
			"Name": name,
		}
		return db.WriteAerospikeMultipleBins(key, bins)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) SetAerospikeConvoMtime(CID string, mtime string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_MTIME_KEY)
	if err == nil {
		// save as bins
		bins := aerospike.BinMap{
			"M_time": mtime,
		}
		return db.WriteAerospikeMultipleBins(key, bins)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) SetAerospikeConvoFiles(CID string, files []string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_FILES_KEY)
	if err == nil {
		// save as bins
		bins := aerospike.BinMap{
			"Files": files,
		}
		return db.WriteAerospikeMultipleBins(key, bins)
	} else {
		ERROR.Println(err)
		return false
	}
}

// deleters
func (db *Database) DeleteAerospikeUser(username string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(username))
	if err == nil {
		return db.DeleteAerospike(key)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) DeleteAerospikeUserActive(username string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_ACTIVE_TABLE, strings.ToUpper(username))
	if err == nil {
		return db.DeleteAerospike(key)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) DeleteAerospikeUserPhone(phonenum string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_PHONES_TABLE, phonenum)
	if err == nil {
		return db.DeleteAerospike(key)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) DeleteAerospikeConvoMembers(CID string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_MEMBERS_KEY)
	if err == nil {
		return db.DeleteAerospike(key)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) DeleteAerospikeConvoName(CID string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_NAME_KEY)
	if err == nil {
		return db.DeleteAerospike(key)
	} else {
		ERROR.Println(err)
		return false
	}
}

func (db *Database) DeleteAerospikeConvoMtime(CID string) bool {
	key, err := aerospike.NewKey(AEROSPIKE_CONVOS_NAMESPACE, CID, AEROSPIKE_CONVOS_MTIME_KEY)
	if err == nil {
		return db.DeleteAerospike(key)
	} else {
		ERROR.Println(err)
		return false
	}
}

// helper for publishing to active web devices
func (db *Database) SendStringToWebDevices(webDevices []string, content string) {
	if db.nats_encodedconn == nil {
		return // don't do anything if nats is nil!
	}
	for _, webz := range webDevices {
		db.nats_encodedconn.Publish(webz, content)
	}
}

// actual Database functions
func (db *Database) CreateUser(data string) UserStruct {
	jsondata := CreateUserCmdStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	// check if user already exists
	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	if storeduser.UsernameUpper != "" {
		// TRACE.Println("user already found in db, don't register!")
		return UserStruct{}
	}
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(jsondata.Password), bcrypt.DefaultCost)
	if err != nil {
		ERROR.Println("Error in bcrypt.GenerateFromPassword()")
		ERROR.Println(err)
	}

	var user = UserStruct{
		Username:      jsondata.Username,
		UsernameUpper: strings.ToUpper(jsondata.Username),
		Password:      string(hashedPassword),
		Email:         jsondata.Email,
		Phone:         jsondata.Phone,
		SecQuests:     jsondata.SecQuests,
		Quota:         2000000, // 2GB
		QuotaUsed:     0,
		ProfilePic:    "https://pingedchat-us1.s3.amazonaws.com/defaultprofile.png",
		// Friends: make([]string, 0), // we use a struct, so we Marshal and Unmarshal into a string here
		// PendingFriends: make([]string, 0), // we use a struct, so we Marshal and Unmarshal into a string here
		CIDs:              "",
		ScheduledMessages: "",
		EmailMtime:        "",
		Android:           make([]string, 0),
		Fireos:            make([]string, 0),
		Ios:               make([]string, 0),
		Web:               make([]string, 1),
	}
	user.Web[0] = jsondata.Token

	// create postgres table for email
	// email table structure:
	// [   FROM_EMAIL   |   TO_EMAILS    |  RECV_EMAIL  |  SUBJECT  |  CONTENT  | ATTACHMENTS |  STARRED  |  UNREAD  |   SPAM   |  DRAFT  |  DELETED  |  RECV_TIME  |    M_TIME   ]
	// [     varchar    |    varchar     |   varchar    |  varchar  |  varchar  |   varchar   |  boolean  |  boolean |  boolean | boolean |  boolean  | timestamptz |  timestamptz ]
	PostgresTableNamestring := `"` + jsondata.Username + `@pinged.email` + `"`
	CurString := "CREATE TABLE " + PostgresTableNamestring + " (from_email varchar NOT NULL, to_emails varchar NOT NULL, recv_email varchar NOT NULL, subject varchar, content varchar, attachments varchar, starred boolean, unread boolean, spam boolean, draft boolean, deleted boolean, recv_time timestamptz NOT NULL, m_time timestamptz NOT NULL, PRIMARY KEY (from_email, recv_time) );"
	// TRACE.Println("CurString: " + CurString)
	if db.postgres_conn != nil {
		_, createerr := db.postgres_conn.Exec(CurString)
		if createerr != nil {
			ERROR.Println("error creating email postgres table in CreateUser: ", createerr)
			// maybe add flag to user struct to show email wasn't created successfully?
		}
	} else {
		// maybe add flag to user struct to show email wasn't created successfully?
	}

	// hash each security question also
	secQuests := user.GetSecurityQuestionStructs()
	for i, e := range secQuests {
		hashedAnswer, err := bcrypt.GenerateFromPassword([]byte(e.Answer), bcrypt.DefaultCost)
		if err != nil {
			ERROR.Println("Error in bcrypt.GenerateFromPassword()")
			ERROR.Println(err)
		}
		secQuests[i].Answer = string(hashedAnswer)
	}
	user.SaveSecurityQuestionStructs(secQuests) // save back to user
	userbins := user.ToAerospikeBins()
	writeSuccess := db.SetAerospikeUser(user.UsernameUpper, userbins)
	if writeSuccess {
		// TRACE.Println("in CreateUser, user = " + user.ToJSONString())
		// link NATS
		if db.nats_encodedconn != nil {
			db.nats_encodedconn.BindRecvChan(jsondata.Token, db.nats_receive)
		}
		return user
	} else {
		return UserStruct{}
	}

}

func (db *Database) ValidateUser(data string) UserStruct {
	jsondata := ValidateUserCmdStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	err := bcrypt.CompareHashAndPassword([]byte(storeduser.Password), []byte(jsondata.Password))
	if err == nil {
		// TRACE.Println("user password match!")
		web_devs := storeduser.Web
		storeduser.Web = append(web_devs, jsondata.Token)
		webBin := aerospike.NewBin("Web", storeduser.Web)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, webBin)
		}
		// link NATS
		if db.nats_encodedconn != nil {
			db.nats_encodedconn.BindRecvChan(jsondata.Token, db.nats_receive)
		}
		// create CID struct using latest data
		userCIDs := storeduser.GetCIDStructs()
		for i, _ := range userCIDs {
			userCIDs[i].M_time = db.GetAerospikeConvoMtime(userCIDs[i].CID)
		}
		storeduser.SaveCIDStructs(userCIDs)
		db.SetAerospikeUser(storeduser.UsernameUpper, storeduser.ToAerospikeBins())
		return storeduser
	} else {
		// TRACE.Println("incorrect user password")
		return UserStruct{}
	}
}

func (db *Database) DeleteUser(data string) string {
	jsondata := ValidateUserCmdStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	// first get user and update all devices that the user has been deleted
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	db.SendStringToWebDevices(user.Web, data)
	// now loop through all pending friends (both incoming and outgoing) and
	// accepted friends and remove user from all lists
	// first loop through incoming friend requests and remove from outgoing friend requests
	userIncomingPendingFriends := user.GetIncomingPendingFriendStructs()
	for _, pendingFriend := range userIncomingPendingFriends {
		friend := db.GetAerospikeUser(strings.ToUpper(pendingFriend.Username))
		friendOutgoingPendingFriends := friend.GetOutgoingPendingFriendStructs()
		for i2, pendingFriend2 := range friendOutgoingPendingFriends {
			if strings.ToUpper(pendingFriend2.Username) == strings.ToUpper(user.UsernameUpper) {
				friendOutgoingPendingFriends = append(friendOutgoingPendingFriends[:i2], friendOutgoingPendingFriends[i2+1:]...)
				friend.SaveOutgoingPendingFriendStructs(friendOutgoingPendingFriends)
				db.SetAerospikeUser(friend.UsernameUpper, friend.ToAerospikeBins()) // save back
				// send update string
				db.SendStringToWebDevices(friend.Web, friend.ToJSONStringWithCmd("UpdateUser"))
				break
			}
		}
	}

	// now loop through outgoing friend requests and delete from friend's incoming requests
	userOutgoingPendingFriends := user.GetOutgoingPendingFriendStructs()
	for _, pendingFriend := range userOutgoingPendingFriends {
		friend := db.GetAerospikeUser(strings.ToUpper(pendingFriend.Username))
		friendIncomingPendingFriends := friend.GetIncomingPendingFriendStructs()
		for i2, pendingFriend2 := range friendIncomingPendingFriends {
			if strings.ToUpper(pendingFriend2.Username) == strings.ToUpper(user.UsernameUpper) {
				friendIncomingPendingFriends = append(friendIncomingPendingFriends[:i2], friendIncomingPendingFriends[i2+1:]...)
				friend.SaveIncomingPendingFriendStructs(friendIncomingPendingFriends)
				db.SetAerospikeUser(friend.UsernameUpper, friend.ToAerospikeBins()) // save back
				// send update string
				db.SendStringToWebDevices(friend.Web, friend.ToJSONStringWithCmd("UpdateUser"))
				break
			}
		}
	}

	// now loop through accepted friends and remove from their friend lists
	userFriends := user.GetFriendStructs()
	for _, acceptedFriend := range userFriends {
		friend := db.GetAerospikeUser(strings.ToUpper(acceptedFriend.Username))
		friendFriends := friend.GetFriendStructs()
		for i2, pendingFriend2 := range friendFriends {
			if strings.ToUpper(pendingFriend2.Username) == strings.ToUpper(user.UsernameUpper) {
				friendFriends = append(friendFriends[:i2], friendFriends[i2+1:]...)
				friend.SaveFriendStructs(friendFriends)
				db.SetAerospikeUser(friend.UsernameUpper, friend.ToAerospikeBins()) // save back
				// send update string
				db.SendStringToWebDevices(friend.Web, friend.ToJSONStringWithCmd("UpdateUser"))
				break
			}
		}
	}

	// now actually delete user
	db.DeleteAerospikeUser(strings.ToUpper(user.UsernameUpper))
	// delete email database
	CurString := `DROP TABLE "` + user.Username + `@pinged.email";`
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil in DeleteUser") // can't do anything with no database connection :(
	}
	_, droperr := db.postgres_conn.Exec(CurString)
	if droperr != nil {
		ERROR.Println("error dropping Postgres table in DeleteUser: ", droperr)
	}
	return data // return DeleteUser, so web app knows to delete user
}

func (db *Database) GetPasswordResetUser(data string) string {
	cmdJSON := CommonCmdStruct{}
	err := json.Unmarshal([]byte(data), &cmdJSON)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	user := db.GetAerospikeUser(strings.ToUpper(cmdJSON.Username))
	if user.Username == "" {
		ERROR.Println("user in GetPasswordResetUser does not exist")
		return data // just send back what we got
	}
	// TRACE.Println("user.SecQuests = " + user.SecQuests)
	secQuests := user.GetSecurityQuestionStructs()
	retuser := PasswordResetUserStruct{}
	for _, question := range secQuests {
		retuser.Questions = append(retuser.Questions, question.Question)
	}
	// TRACE.Println("questions in GetPasswordResetUser:")
	// TRACE.Println(retuser.Questions)
	retuser.Cmd = "GetPasswordResetUser" // return same as incoming command
	retuser.Username = cmdJSON.Username
	retUserString, err := ffjson.Marshal(retuser)
	if err != nil {
		ERROR.Println("Error in ffjson.Marshal(retuser)")
		ERROR.Println(err)
	}
	// TRACE.Println("retUserString = "  + string(retUserString))
	return string(retUserString)
}

func (db *Database) ResetUserPassword(data string) string {
	// TRACE.Println("in ResetUserPassword")
	jsondata := CreateUserCmdStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	// check if user already exists
	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	if storeduser.UsernameUpper == "" {
		// TRACE.Println("user not found, returning from ResetUserPassword")
		return `{"cmd":"ResetUserPassword", "success":"false"}`
	}
	if len(jsondata.Password) < 6 {
		return `{"cmd":"ResetUserPassword", "success":"false"}`
	}
	// compare answers
	// TRACE.Println("storeduser.SecQuests = " + storeduser.SecQuests)
	storedSecQuests := storeduser.GetSecurityQuestionStructs()
	var passedSecQuests []SecurityQuestionStruct
	err := json.Unmarshal([]byte(jsondata.SecQuests), &passedSecQuests)
	if err != nil {
		ERROR.Println("error in json.Unmarshal, jsondata.SecQuests = " + jsondata.SecQuests)
		ERROR.Println(err)
		passedSecQuests = make([]SecurityQuestionStruct, 0)
	}
	pass1, pass2 := false, false
	for _, storedQuestion := range storedSecQuests {
		for _, passedQuestion := range passedSecQuests {
			if storedQuestion.Question == passedQuestion.Question {
				// TRACE.Println("comparing security questions " + string(storedQuestion.Question) + " and " + string(passedQuestion.Question))
				// TRACE.Println("comparing security answers " + string(storedQuestion.Answer) + " and " + string(passedQuestion.Answer))
				//if storedQuestion.Answer == passedQuestion.Answer {
				err := bcrypt.CompareHashAndPassword([]byte(storedQuestion.Answer), []byte(passedQuestion.Answer))
				if err == nil { // security answers match!
					if pass1 == false {
						pass1 = true
						// TRACE.Println("setting pass1 = true")
					} else {
						pass2 = true
						// TRACE.Println("setting pass2 = true")
						break // at this point, both security answers have been answered
					}
				}
			}
		}
	}
	if pass1 && pass2 {
		// both security questions passed
		// set password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(jsondata.Password), bcrypt.DefaultCost)
		if err != nil {
			ERROR.Println("Error in bcrypt.GenerateFromPassword()")
			ERROR.Println(err)
			return `{"cmd":"ResetUserPassword", "success":"false"}` // no password update
		}
		// TRACE.Println("setting new hashed password to " + storeduser.Username)
		storeduser.Password = string(hashedPassword)
		passwordBin := aerospike.NewBin("Password", storeduser.Password)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, passwordBin)
		}
		// return success
		return `{"cmd":"ResetUserPassword", "success":"true"}`
	}
	// if here, then we didn't successfully complete the password reset above
	return `{"cmd":"ResetUserPassword", "success":"false"}`
}

func (db *Database) ChangeUserPassword(data string) string {
	jsondata := ChangeUserPasswordStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// compare old password first
	err2 := bcrypt.CompareHashAndPassword([]byte(storeduser.Password), []byte(jsondata.OldPassword))
	if err2 == nil {
		// passwords match!  now set new password.
		hashedPassword, err3 := bcrypt.GenerateFromPassword([]byte(jsondata.NewPassword), bcrypt.DefaultCost)
		if err2 != nil {
			ERROR.Println("Error in bcrypt.GenerateFromPassword()")
			ERROR.Println(err3)
			return `{"cmd":"ResetUserPassword", "success":"false"}` // no password update
		}
		// TRACE.Println("setting new hashed password to " + storeduser.Username)
		storeduser.Password = string(hashedPassword)
		passwordBin := aerospike.NewBin("Password", storeduser.Password)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, passwordBin)
		}
		// return success
		return `{"cmd":"ChangeUserPassword", "success":"true"}`
	} else {
		return `{"cmd":"ChangeUserPassword", "success":"false"}`
	}
	// if here, we were not successful somewhere
	return `{"cmd":"ChangeUserPassword", "success":"false"}`
}

func (db *Database) AddToQuota(data string) string {
	jsondata := QuotaCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	// get stored user
	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// add to quota
	storeduser.Quota += jsondata.Quota
	// save back
	QuotaUsedBin := aerospike.NewBin("Quota", int(storeduser.Quota))
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, QuotaUsedBin)
	}
	// send our string we received to all devices, they'll add it in themselves easily enough
	db.SendStringToWebDevices(storeduser.Web, data) // send to all web devices
	return ""                                       // we sent to web devices above, so don't send twice
}

func (db *Database) AddToQuotaUsed(data string) string {
	jsondata := QuotaCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	// get stored user
	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	storeduser.QuotaUsed += jsondata.QuotaUsed
	if storeduser.QuotaUsed > storeduser.Quota {
		return `{"cmd":"AddToQuotaUsed", "message":"You are over your storage quota!  Please add more space to keep uploading media."}`
	} else if storeduser.QuotaUsed == storeduser.Quota {
		return `{"cmd":"AddToQuotaUsed", "message":"You have reached your storage quota!  Please add more space to keep uploading media."}`
	}
	// save back
	QuotaUsedBin := aerospike.NewBin("QuotaUsed", int(storeduser.QuotaUsed))
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, QuotaUsedBin)
	}
	// send our string we received to all devices, they'll add it in themselves easily enough
	db.SendStringToWebDevices(storeduser.Web, data) // send to all web devices
	return ""                                       // we sent to web devices above, so don't send twice
}

func (db *Database) AddScheduledMessage(data string) string {
	jsondata := ScheduledMessagesCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	// get stored user
	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// create new scheduled message and append
	newScheduledMessage := ScheduledMessagesStruct{
		CID:     jsondata.CID,
		Time:    jsondata.Time,
		Content: jsondata.Content,
	}
	userScheduledMessages := storeduser.GetScheduledMessagesStructs()
	userScheduledMessages = append(userScheduledMessages, newScheduledMessage)
	storeduser.SaveScheduledMessagesStructs(userScheduledMessages)
	// now save user's scheduled messages back
	ScheduledMessagesBin := aerospike.NewBin("SchedMessages", storeduser.ScheduledMessages)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, ScheduledMessagesBin)
	}
	// send our string we received to all devices, they'll add it in themselves easily enough
	db.SendStringToWebDevices(storeduser.Web, data) // send to all web devices

	// now we'll add to new postgres table
	CurString := "INSERT INTO " + POSTGRES_SCHEDULED_MESSAGES_TABLE + " (CID, f_username, content, m_time) VALUES ($1, $2, $3, $4)" // conversation ID is the table name
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil in AddScheduledMessage") // can't do anything with no database connection :(
	}
	_, inserterr := db.postgres_conn.Exec(CurString, jsondata.CID, jsondata.Username, jsondata.Content, jsondata.Time)
	if inserterr != nil {
		ERROR.Println("error adding to Postgres table in AddScheduledMessage: ", inserterr)
		return ""
	}
	return "" // we sent to web devices above, so don't send twice
}

func (db *Database) RemoveScheduledMessage(data string) string {
	// TRACE.Println("removing user " + username + " android device " + android)
	jsondata := ScheduledMessagesCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	userScheduledMessages := storeduser.GetScheduledMessagesStructs()
	// loop through web devices, if web device is not the token add to slice
	// then save back to the user
	for i, e := range userScheduledMessages {
		TRACE.Println("e.Time = " + e.Time + " , jsondata.Time = " + jsondata.Time)
		eTime, _ := time.Parse(time.RFC3339Nano, e.Time)
		jsonTime, _ := time.Parse(time.RFC3339Nano, jsondata.Time)
		TRACE.Println("eTime.Equal(jsonTime) : ")
		TRACE.Println(eTime.Equal(jsonTime))
		if eTime.Equal(jsonTime) && e.CID == jsondata.CID && e.Content == jsondata.Content {
			userScheduledMessages = append(userScheduledMessages[:i], userScheduledMessages[i+1:]...) // splice out
		}
	}
	storeduser.SaveScheduledMessagesStructs(userScheduledMessages)
	TRACE.Println("in RemoveScheduledMessage, storeduser.ScheduledMessages = ")
	TRACE.Println(storeduser.ScheduledMessages)
	ScheduledMessagesBin := aerospike.NewBin("SchedMessages", storeduser.ScheduledMessages)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, ScheduledMessagesBin)
	}
	retstr := storeduser.ToJSONStringWithCmd("RemoveScheduledMessage")
	db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices

	// now remove from db if there
	CurString := `DELETE FROM ` + POSTGRES_SCHEDULED_MESSAGES_TABLE + ` WHERE CID = $1 AND m_time = $2 AND f_username = $3 AND content = $4;`
	TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("in handleScheduledMessages, db.postgres_conn == nil , so returning :(")
		return "" // can't do anything with no database connection :(
	}
	m_time, _ := time.Parse(time.RFC3339Nano, jsondata.Time)
	_, deleteErr := db.postgres_conn.Exec(CurString, jsondata.CID, m_time, jsondata.Username, jsondata.Content)
	if deleteErr, ok := deleteErr.(*pq.Error); ok {
		ERROR.Println("pq delete error in RemoveScheduledMessage:", deleteErr.Code.Name())
		ERROR.Println(deleteErr)
	}
	return "" // we sent to web devices above, so don't send twice
}

func (db *Database) RemoveAllScheduledMessages(data string) string {
	// TRACE.Println("removing user " + username + " android device " + android)
	jsondata := ScheduledMessagesCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	storeduser.ScheduledMessages = "" // simply clear them out like this
	ScheduledMessagesBin := aerospike.NewBin("SchedMessages", storeduser.ScheduledMessages)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, ScheduledMessagesBin)
	}
	retstr := `{"cmd":"RemoveAllScheduledMessages"}`
	db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices

	// now remove from db if there
	CurString := `DELETE FROM ` + POSTGRES_SCHEDULED_MESSAGES_TABLE + ` WHERE f_username = $1;`
	TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("in handleScheduledMessages, db.postgres_conn == nil , so returning :(")
		return "" // can't do anything with no database connection :(
	}
	_, deleteErr := db.postgres_conn.Exec(CurString, jsondata.Username)
	if deleteErr, ok := deleteErr.(*pq.Error); ok {
		ERROR.Println("pq delete error in RemoveAllScheduledMessage:", deleteErr.Code.Name())
		ERROR.Println(deleteErr)
	}
	return "" // we sent to web devices above, so don't send twice
}

func (db *Database) AddAndroidDev(data string) string {
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// first make sure that the device is not already stored
	dev_found := false
	for _, e := range storeduser.Android {
		if e == jsondata.Device {
			dev_found = true
			break
		}
	}
	if !dev_found {
		TRACE.Println("adding android device : " + jsondata.Device)
		// device is not added to user, so append and save
		storeduser.Android = append(storeduser.Android, jsondata.Device)
		androidBin := aerospike.NewBin("Android", storeduser.Android)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, androidBin)
		}
		retstr := storeduser.ToJSONStringWithCmd("AddAndroidDev")
		db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
		return ""                                                 // we sent to web devices above, so don't send twice
	}
	return ""
}

func (db *Database) ChangeAndroidDev(data string) string {
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// first, remove old device
	for i, e := range storeduser.Android {
		if e == jsondata.OldDevice {
			storeduser.Android = append(storeduser.Android[:i], storeduser.Android[i+1:]...) // splice out
			break
		}
	}
	// now make sure that the new device is not already stored
	dev_found := false
	for _, e := range storeduser.Android {
		if e == jsondata.Device {
			dev_found = true
			break
		}
	}
	if !dev_found {
		TRACE.Println("adding android device : " + jsondata.Device)
		// device is not added to user, so append and save
		storeduser.Android = append(storeduser.Android, jsondata.Device)
		androidBin := aerospike.NewBin("Android", storeduser.Android)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, androidBin)
		}
		retstr := storeduser.ToJSONStringWithCmd("AddAndroidDev")
		db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
		return ""                                                 // we sent to web devices above, so don't send twice
	}
	return ""
}

func (db *Database) RemoveAndroidDev(data string) string {
	// TRACE.Println("removing user " + username + " android device " + android)
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	android_devs := storeduser.Android
	// loop through web devices, if web device is not the token add to slice
	// then save back to the user
	for i, e := range android_devs {
		if e == jsondata.Device {
			storeduser.Android = append(storeduser.Android[:i], storeduser.Android[i+1:]...) // splice out
		}
	}
	TRACE.Println("in RemoveAndroidDev, storeduser.Android = ")
	TRACE.Println(storeduser.Android)
	androidBin := aerospike.NewBin("Android", storeduser.Android)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, androidBin)
	}
	retstr := storeduser.ToJSONStringWithCmd("RemoveAndroidDev")
	db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
	return ""                                                 // we sent to web devices above, so don't send twice
}

func (db *Database) AddIosDev(data string) string {
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// first make sure that the device is not already stored
	dev_found := false
	for _, e := range storeduser.Ios {
		if e == jsondata.Device {
			dev_found = true
			break
		}
	}
	if !dev_found {
		// device is not added to user, so append and save
		storeduser.Ios = append(storeduser.Ios, jsondata.Device)
		iosBin := aerospike.NewBin("Ios", storeduser.Ios)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, iosBin)
		}
		retstr := storeduser.ToJSONStringWithCmd("AddIosDev")
		db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
		return ""                                                 // we sent to web devices above, so don't send twice
	}
	return ""
}

func (db *Database) ChangeIosDev(data string) string {
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// first, remove old device
	for i, e := range storeduser.Ios {
		if e == jsondata.Device {
			storeduser.Ios = append(storeduser.Ios[:i], storeduser.Ios[i+1:]...) // splice out
		}
	}
	// now make sure that the new device is not already stored
	dev_found := false
	for _, e := range storeduser.Ios {
		if e == jsondata.Device {
			dev_found = true
			break
		}
	}
	if !dev_found {
		// device is not added to user, so append and save
		storeduser.Ios = append(storeduser.Ios, jsondata.Device)
		iosBin := aerospike.NewBin("Ios", storeduser.Ios)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, iosBin)
		}
		retstr := storeduser.ToJSONStringWithCmd("AddIosDev")
		db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
		return ""                                                 // we sent to web devices above, so don't send twice
	}
	return ""
}

func (db *Database) RemoveIosDev(data string) string {
	// TRACE.Println("removing user " + username + " ios device " + ios)
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	ios_devs := storeduser.Ios
	// loop through web devices, if web device is not the token add to slice
	// then save back to the user
	for i, e := range ios_devs {
		if e == jsondata.Device {
			storeduser.Ios = append(storeduser.Ios[:i], storeduser.Ios[i+1:]...) // splice out
		}
	}
	iosBin := aerospike.NewBin("Ios", storeduser.Ios)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, iosBin)
	}
	retstr := storeduser.ToJSONStringWithCmd("RemoveIosDev")
	db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
	return ""                                                 // we sent to web devices above, so don't send twice
}

func (db *Database) AddFireosDev(data string) string {
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// first make sure that the device is not already stored
	dev_found := false
	for _, e := range storeduser.Fireos {
		if e == jsondata.Device {
			dev_found = true
			break
		}
	}
	if !dev_found {
		// device is not added to user, so append and save
		storeduser.Fireos = append(storeduser.Fireos, jsondata.Device)
		fireosBin := aerospike.NewBin("Fireos", storeduser.Fireos)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, fireosBin)
		}
		retstr := storeduser.ToJSONStringWithCmd("AddFireosDev")
		db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
		return ""                                                 // we sent to web devices above, so don't send twice
	}
	return ""
}

func (db *Database) ChangeFireosDev(data string) string {
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	// first, remove old device
	for i, e := range storeduser.Fireos {
		if e == jsondata.Device {
			storeduser.Fireos = append(storeduser.Fireos[:i], storeduser.Fireos[i+1:]...) // splice out
		}
	}
	// now make sure that the device is not already stored
	dev_found := false
	for _, e := range storeduser.Fireos {
		if e == jsondata.Device {
			dev_found = true
			break
		}
	}
	if !dev_found {
		// device is not added to user, so append and save
		storeduser.Fireos = append(storeduser.Fireos, jsondata.Device)
		fireosBin := aerospike.NewBin("Fireos", storeduser.Fireos)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, fireosBin)
		}
		retstr := storeduser.ToJSONStringWithCmd("AddFireosDev")
		db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
		return ""                                                 // we sent to web devices above, so don't send twice
	}
	return ""
}

func (db *Database) RemoveFireosDev(data string) string {
	// TRACE.Println("removing user " + username + " fireos device " + fireos)
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	fireos_devs := storeduser.Fireos
	// loop through web devices, if web device is not the token add to slice
	// then save back to the user
	for i, e := range fireos_devs {
		if e == jsondata.Device {
			storeduser.Fireos = append(storeduser.Fireos[:i], storeduser.Fireos[i+1:]...) // splice out
		}
	}
	fireosBin := aerospike.NewBin("Fireos", storeduser.Fireos)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, fireosBin)
	}
	retstr := storeduser.ToJSONStringWithCmd("RemoveFireosDev")
	db.SendStringToWebDevices(storeduser.Web, string(retstr)) // send to all web devices
	return ""                                                 // we sent to web devices above, so don't send twice
}

func (db *Database) RemoveUserWebToken(username string, token string) string {
	// TRACE.Println("removing user " + username + " web token " + token)
	if strings.TrimSpace(username) == "" || strings.TrimSpace(token) == "" {
		return ""
	}
	storeduser := db.GetAerospikeUser(strings.ToUpper(username))
	// check if length of web devices is > 0
	if len(storeduser.Web) < 1 {
		return "" // can't splice nothing
	} else if len(storeduser.Web) == 1 {
		// we know there's only the one web device, and we most likely lost internet connection, so clear the array
		storeduser.Web = make([]string, 0) // empty array
	} else {
		// user has multiple devices online simultaneously, so loop through web devices
		for i, e := range storeduser.Web {
			if e == token || e == "" { // we remove if it equals the token, or if empty string
				if len(storeduser.Web) <= (i + 1) {
					storeduser.Web = storeduser.Web[:i] // take everything up to this point
				} else {
					storeduser.Web = append(storeduser.Web[:i], storeduser.Web[i+1:]...) // splice out
				}
				break
			}
		}
	}
	// then save back to the user
	webBin := aerospike.NewBin("Web", storeduser.Web)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, webBin)
	}
	return storeduser.ToJSONStringWithCmd("RemoveWebDev")
}

func (db *Database) RemoveWebDev(data string) string {
	jsondata := ChangeDeviceStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	return db.RemoveUserWebToken(jsondata.Username, jsondata.Device)
}

// conversations
func (db *Database) CreateConversation(data string) string {
	// TRACE.Println("data in CreateConversation = " + data)
	jsondata := CIDCommandStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	CID := uuid.NewV4()
	CIDstring := uuid.Formatter(CID, uuid.CleanHyphen)
	// TRACE.Println("new generated CID: " + CIDstring)
	// TRACE.Println("jsondata in CreateConversation:")
	// TRACE.Println(jsondata)
	PostgresCIDstring := `"` + CIDstring + `"`
	CurString := "CREATE TABLE " + PostgresCIDstring + " (f_username varchar NOT NULL, m_time timestamptz NOT NULL, content varchar NOT NULL, PRIMARY KEY (f_username, m_time) );"
	// TRACE.Println("CurString: " + CurString)
	if db.postgres_conn == nil {
		return "" // can't do anything with no database connection :(
	}
	_, createerr := db.postgres_conn.Exec(CurString)
	if createerr != nil {
		ERROR.Println("error:", createerr)
		return ""
	}
	// save to new CID in aerospike
	var memberArray ConvoMemberArray
	// loop through members
	for _, m := range jsondata.Members {
		memberArray = append(memberArray, ConvoMember{Username: string(m), ReadTime: string(jsondata.M_time), Typing: false})
		// add CID to each member
		db.AddConversationToUser(CIDstring, jsondata.Name, string(m), jsondata.M_time) // add CID to user
	}
	db.SetAerospikeConvoMembers(CIDstring, memberArray.ToStringArray())
	db.SetAerospikeConvoName(CIDstring, jsondata.Name)
	db.SetAerospikeConvoMtime(CIDstring, jsondata.M_time)

	// send message out to each user with new CID and all data
	jsondata.CID = CIDstring
	jsondata.Members = memberArray.ToStringArray()

	datastr, err := ffjson.Marshal(jsondata)
	if err != nil {
		ERROR.Println("err in ffjson.Marshal(jsondata) in CreateConversation:")
		ERROR.Println(err)
		return "" // return on error :(
	}
	for _, m := range memberArray {
		member := db.GetAerospikeUser(m.Username)
		db.SendStringToWebDevices(member.Web, string(datastr))
	}

	return "" // return nothing, it should already be sent above in SendStringToWebDevices
}

func (db *Database) AddConversationToUser(newCID string, name string, username string, m_time string) {
	// TRACE.Println("adding CID " + newCID + " named " + name + " to username " + username + " at m_time " + m_time)
	storeduser := db.GetAerospikeUser(strings.ToUpper(username))
	CIDFound := false
	storedCIDs := storeduser.GetCIDStructs()
	for i, e := range storedCIDs {
		if e.CID == newCID {
			// if the user already is a part of the conversation, update the name and m_time
			// storeduser.CIDs[i].Name = name
			storedCIDs[i].M_time = m_time
			CIDFound = true
			break
		}
	}
	if !CIDFound {
		newCID := UserCIDStruct{
			// Name: name,
			CID:         newCID,
			M_time:      m_time,
			UnreadCount: 0,
		}
		storedCIDs = append(storedCIDs, newCID)
	}
	// TRACE.Println("storedCIDs:")
	// TRACE.Println(storedCIDs)
	storeduser.SaveCIDStructs(storedCIDs)
	// TRACE.Println("user after saveCIDStructs: " + storeduser.ToJSONString())
	// save back
	CIDBin := aerospike.NewBin("CIDs", storeduser.CIDs)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, CIDBin)
	}
}

func (db *Database) AddUsersToConversation(data string) string {
	// jsondata.Members are new members
	jsondata := CIDCommandStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	convoMembersStrings := db.GetAerospikeConvoMembers(jsondata.CID)
	convoMembers := ToConvoMemberArray(convoMembersStrings)

	newUserAdded := false
	for _, newUser := range jsondata.Members {
		TRACE.Println("newUser = " + newUser)
		userInConversation := false
		for _, m := range convoMembers {
			if strings.ToUpper(m.Username) == strings.ToUpper(jsondata.Username) {
				userInConversation = true
				break
			} // end if strings.ToUpper(m.Username) == strings.ToUpper(jsondata.Username)
		} // end for _, m := range convoMembers

		// only add to conversation if user is not already in conversation
		if !userInConversation {
			TRACE.Println(newUser + " was not in conversation, adding")
			newMember := ConvoMember{
				Username: newUser,
				ReadTime: jsondata.M_time,
				Typing:   false,
			}
			convoMembers = append(convoMembers, newMember)
			sort.Sort(convoMembers) // alphabetize by Username
			newUserAdded = true
			// append new CID to user struct
			user := db.GetAerospikeUser(newUser)
			userCIDs := user.GetCIDStructs()
			newCID := UserCIDStruct{
				CID:         jsondata.CID,
				M_time:      jsondata.M_time,
				UnreadCount: 0,
			}
			userCIDs = append(userCIDs, newCID)
			TRACE.Println("new userCIDs:")
			TRACE.Println(userCIDs)
			user.SaveCIDStructs(userCIDs)
			CIDBin := aerospike.NewBin("CIDs", user.CIDs)
			key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
			if err == nil {
				db.UpdateAerospikeSingleBin(key, CIDBin)
			}
		} // end if !userInConversation
	} // end for _, newUser := range jsondata.Members

	if newUserAdded {
		// save to general CID Aerospike space
		db.SetAerospikeConvoMembers(jsondata.CID, convoMembers.ToStringArray())
		// update time of convo
		db.SetAerospikeConvoMtime(jsondata.CID, jsondata.M_time)

		membersString, err := ffjson.Marshal(convoMembers)
		if err != nil {
			ERROR.Println("Error in ffjson.Marshal(convoMembers) in RemoveUserFromConversation")
			ERROR.Println(err)
		}

		// loop through users and send message if online saying to update information
		updateString := `{"cmd":"AddUsersToConversation", "CID":"` + jsondata.CID + `", "M_time:":"` + jsondata.M_time + `", "Members":` + string(membersString) + `}`
		for _, m := range convoMembers {
			user := db.GetAerospikeUser(m.Username) // maybe use GetAerospikeUserActive later
			for _, webz := range user.Web {
				if db.nats_encodedconn != nil {
					db.nats_encodedconn.Publish(webz, string(updateString))
				}
			}
		}
	}

	return ""
}

func (db *Database) RemoveUserFromConversation(data string) string {
	jsondata := CIDCommandStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	convoMembersStrings := db.GetAerospikeConvoMembers(jsondata.CID)
	convoMembers := ToConvoMemberArray(convoMembersStrings)

	userRemoved := false
	for i, m := range convoMembers {
		if strings.ToUpper(m.Username) == strings.ToUpper(jsondata.Username) {
			convoMembers = append(convoMembers[:i], convoMembers[i+1:]...) // splice out
			db.SetAerospikeConvoMembers(jsondata.CID, convoMembers.ToStringArray())
			userRemoved = true
			break
		}
	}

	if userRemoved {
		// update time of convo
		db.SetAerospikeConvoMtime(jsondata.CID, jsondata.M_time)
		// loop through users and send message if online saying to update information
		membersString, err := ffjson.Marshal(convoMembers)
		if err != nil {
			ERROR.Println("Error in ffjson.Marshal(convoMembers) in RemoveUserFromConversation")
			ERROR.Println(err)
		}
		updateString := `{"cmd":"update_convo_members", "CID":"` + jsondata.CID + `", "Members":` + string(membersString) + `}`
		for _, m := range convoMembers {
			user := db.GetAerospikeUser(m.Username) // maybe use GetAerospikeUserActive later
			for _, webz := range user.Web {
				if db.nats_encodedconn != nil {
					db.nats_encodedconn.Publish(webz, string(updateString))
				}
			}
		}

		// now update user struct
		user := db.GetAerospikeUser(jsondata.Username)
		userCIDs := user.GetCIDStructs()
		for i, CID := range userCIDs {
			if CID.CID == jsondata.CID {
				// splice out
				userCIDs = append(userCIDs[:i], userCIDs[i+1:]...) // splice out
				break                                              // only one CID removed
			}
		}
		user.SaveCIDStructs(userCIDs)
		// TRACE.Println("saving user: " + user.ToJSONString())
		CIDBin := aerospike.NewBin("CIDs", user.CIDs)
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, CIDBin)
			return user.ToJSONStringWithCmd("RemoveUserFromConversation")
		}
	}

	return ""
}

func (db *Database) ChangeConvoName(data string) string {
	jsondata := CIDCommandStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	db.SetAerospikeConvoName(jsondata.CID, jsondata.Name)

	convoMembersStrings := db.GetAerospikeConvoMembers(jsondata.CID)
	convoMembers := ToConvoMemberArray(convoMembersStrings)

	updateString := `{"cmd":"update_convo_name", "CID":"` + jsondata.CID + `", "name":"` + jsondata.Name + `"}`
	// loop through users and send message if online saying to update information
	for _, m := range convoMembers {
		user := db.GetAerospikeUser(m.Username) // maybe use GetAerospikeUserActive later
		for _, webz := range user.Web {
			if db.nats_encodedconn != nil {
				db.nats_encodedconn.Publish(webz, string(updateString))
			}
		}
	}

	return ""
}

func (db *Database) UpdateConvoFiles(data string) string {
	jsondata := CmdConvoAddFileStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	curFilesStringArr := db.GetAerospikeConvoFiles(jsondata.CID)
	curFilesStructArr := ToConvoFileListArray(curFilesStringArr)

	// loop through and see if filename exists, if it does, delete it and re-add it
	for i, e := range curFilesStructArr {
		if e.FileURL == jsondata.FileURL {
			TRACE.Println("FileURL matching in UpdateConvoFiles, so removing original")
			curFilesStructArr = append(curFilesStructArr[:i], curFilesStructArr[i+1:]...) // splice out
			break                                                                         // should only happen once, we can optimize and skip the rest :) OPTIMIZE PRIME
		}
	}

	newFile := ConvoFileListStruct{
		FromUsername: jsondata.FromUsername,
		FileURL:      jsondata.FileURL,
		M_time:       jsondata.M_time,
	}

	// add newFile to array and save back
	curFilesStructArr = append(curFilesStructArr, newFile)
	db.SetAerospikeConvoFiles(jsondata.CID, curFilesStructArr.ToStringArray())

	// send to every user in the group to update their CID files
	convoMembersStrings := db.GetAerospikeConvoMembers(jsondata.CID)
	convoMembers := ToConvoMemberArray(convoMembersStrings)
	if db.nats_encodedconn != nil {
		// TRACE.Println("looping through each member and publishing user status update to them")
		for _, e := range convoMembers {
			recipient := db.GetAerospikeUser(strings.ToUpper(e.Username))
			db.SendStringToWebDevices(recipient.Web, data)
		}
	}

	return ""
}

func (db *Database) GetConvoData(data string) string {
	jsondata := CIDCommandStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	// TRACE.Println("jsondata in GetConvoData:")
	// TRACE.Println(jsondata)
	var rows *sql.Rows
	var err error
	retMessages := make([]ConvoRowStruct, 0)
	PostgresCIDstring := `"` + jsondata.CID + `"`
	// TRACE.Println("PostgresCIDsstring = " + string(PostgresCIDstring))
	if db.postgres_conn == nil {
		return "" // can't do anything with no database connection :(
	}
	if jsondata.M_time != "" {
		// use m_time to get most recent messages
		// TRACE.Println("Using m_time to get most recent messages from postgres")
		// need to put the timestamp in single quotes
		// VALID:
		// SELECT f_username, content, m_time FROM  "4730d9e6-b719-411a-b179-62f35528c7d5" WHERE m_time > '2015-06-12T19:16:29.119Z'::timestamptz
		rows, err = db.postgres_conn.Query(`SELECT f_username, content, m_time FROM ` + PostgresCIDstring + ` WHERE m_time > '` + jsondata.M_time + `'::timestamptz ORDER BY m_time ASC;`) // conversation ID is the table name, return 20 latest messages
	} else {
		// no m_time, so get 20 most recent messages
		// TRACE.Println("getting most recent 20 messages from postgres")
		// Average ms for 20 rows: 497ms
		// Average ms for 50 rows: 469.25ms // I KNOW, RIGHT???
		PostgresString := `WITH results AS (SELECT f_username, content, m_time FROM ` + PostgresCIDstring + ` ORDER BY m_time DESC LIMIT 50) SELECT * FROM results ORDER BY m_time ASC;`
		// TRACE.Println("PostgresString = " + string(PostgresString))
		rows, err = db.postgres_conn.Query(PostgresString)
	}
	if err, ok := err.(*pq.Error); ok {
		ERROR.Println("pq error:", err.Code.Name())
		ERROR.Println(err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var SQLF_username string
			var SQLContent string
			var SQLM_time time.Time
			if err := rows.Scan(&SQLF_username, &SQLContent, &SQLM_time); err != nil {
				ERROR.Println(err)
			}
			rowStruct := ConvoRowStruct{
				F_username: SQLF_username,
				Content:    SQLContent,
				M_time:     SQLM_time,
			}
			// convoRow, err := ffjson.Marshal(rowStruct)
			if err != nil {
				ERROR.Println("Error in ffjson.Marshal(rowStruct)")
				ERROR.Println(err)
			}
			// TRACE.Println("convoRow: " + string(convoRow))
			retMessages = append(retMessages, rowStruct)
		}
	}

	// now get from Aerospike
	retName := db.GetAerospikeConvoName(jsondata.CID)
	retMtime := db.GetAerospikeConvoMtime(jsondata.CID)
	retMembers := db.GetAerospikeConvoMembers(jsondata.CID)
	retFiles := db.GetAerospikeConvoFiles(jsondata.CID)

	retCmd := ConvoDataStruct{
		Cmd:      "GetConvoData",
		CID:      jsondata.CID,
		Name:     retName,
		M_time:   retMtime,
		Members:  retMembers,
		Files:    retFiles,
		Messages: retMessages,
	}
	retCmdString, err := ffjson.Marshal(retCmd)
	if err != nil {
		ERROR.Println("Error in ffjson.Marshal(retCmd)")
		ERROR.Println(err)
	}
	// TRACE.Println("GetConvoData returning " + string(retCmdString))
	return string(retCmdString)
}

func (db *Database) GetAllConvoData(data string) string {
	jsondata := CIDCommandStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	// TRACE.Println("jsondata in GetConvoData:")
	// TRACE.Println(jsondata)
	var rows *sql.Rows
	var err error
	retMessages := make([]ConvoRowStruct, 0)
	PostgresCIDstring := `"` + jsondata.CID + `"`
	// TRACE.Println("PostgresCIDsstring = " + string(PostgresCIDstring))
	if db.postgres_conn == nil {
		return "" // can't do anything with no database connection :(
	}

	rows, err = db.postgres_conn.Query(`SELECT f_username, content, m_time FROM ` + PostgresCIDstring + `;`) // conversation ID is the table name, return 20 latest messages
	if err, ok := err.(*pq.Error); ok {
		ERROR.Println("pq error:", err.Code.Name())
		ERROR.Println(err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var SQLF_username string
			var SQLContent string
			var SQLM_time time.Time
			if err := rows.Scan(&SQLF_username, &SQLContent, &SQLM_time); err != nil {
				ERROR.Println(err)
			}
			rowStruct := ConvoRowStruct{
				F_username: SQLF_username,
				Content:    SQLContent,
				M_time:     SQLM_time,
			}
			// convoRow, err := ffjson.Marshal(rowStruct)
			if err != nil {
				ERROR.Println("Error in ffjson.Marshal(rowStruct)")
				ERROR.Println(err)
			}
			// TRACE.Println("convoRow: " + string(convoRow))
			retMessages = append(retMessages, rowStruct)
		}
	}

	// now get from Aerospike
	retName := db.GetAerospikeConvoName(jsondata.CID)
	retMtime := db.GetAerospikeConvoMtime(jsondata.CID)
	retMembers := db.GetAerospikeConvoMembers(jsondata.CID)
	retFiles := db.GetAerospikeConvoFiles(jsondata.CID)

	retCmd := ConvoDataStruct{
		Cmd:      "GetConvoData",
		CID:      jsondata.CID,
		Name:     retName,
		M_time:   retMtime,
		Members:  retMembers,
		Files:    retFiles,
		Messages: retMessages,
	}
	retCmdString, err := ffjson.Marshal(retCmd)
	if err != nil {
		ERROR.Println("Error in ffjson.Marshal(retCmd)")
		ERROR.Println(err)
	}
	// TRACE.Println("GetConvoData returning " + string(retCmdString))
	return string(retCmdString)
}

func (db *Database) GetMoreConvoMessages(data string) string {
	jsondata := CIDCommandStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	// TRACE.Println("jsondata in GetConvoData:")
	// TRACE.Println(jsondata)
	var rows *sql.Rows
	var err error
	retMessages := make([]ConvoRowStruct, 0)
	PostgresCIDstring := `"` + jsondata.CID + `"`
	// TRACE.Println("PostgresCIDsstring = " + string(PostgresCIDstring))
	if db.postgres_conn == nil {
		return "" // can't do anything with no database connection :(
	}
	if jsondata.M_time == "" {
		return "" // we need an M_time to get more messages from some point in time
	}
	// in GetConvoData, I bulk add to the javascript db so I return in order of earliest -> latest
	// but here, I want to push to the top each message, so I return in order of lastest -> earliest
	// which when pushed individually, keeps the whole list earliest -> latest
	// the javascript db sorts by m_time, so insertion order isn't big factor
	rows, err = db.postgres_conn.Query(`SELECT f_username, content, m_time FROM ` + PostgresCIDstring + ` WHERE m_time < '` + jsondata.M_time + `'::timestamptz ORDER BY m_time DESC LIMIT 50;`)
	if err, ok := err.(*pq.Error); ok {
		ERROR.Println("pq error:", err.Code.Name())
		ERROR.Println(err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var SQLF_username string
			var SQLContent string
			var SQLM_time time.Time
			if err := rows.Scan(&SQLF_username, &SQLContent, &SQLM_time); err != nil {
				ERROR.Println(err)
			}
			rowStruct := ConvoRowStruct{
				F_username: SQLF_username,
				Content:    SQLContent,
				M_time:     SQLM_time,
			}
			// convoRow, err := ffjson.Marshal(rowStruct)
			if err != nil {
				ERROR.Println("Error in ffjson.Marshal(rowStruct)")
				ERROR.Println(err)
			}
			// TRACE.Println("convoRow: " + string(convoRow))
			retMessages = append(retMessages, rowStruct)
		}
	}

	retCmd := ConvoDataStruct{
		Cmd:      "GetMoreConvoMessages",
		CID:      jsondata.CID,
		Messages: retMessages,
	}
	retCmdString, err := ffjson.Marshal(retCmd)
	if err != nil {
		ERROR.Println("Error in ffjson.Marshal(retCmd)")
		ERROR.Println(err)
	}
	// TRACE.Println("GetConvoData returning " + string(retCmdString))
	return string(retCmdString)
}

// EMAILS
func (db *Database) GetAllEmails(data string) string {
	jsondata := CIDCommandStruct{}
	json.Unmarshal([]byte(data), &jsondata)
	// TRACE.Println("jsondata in GetAllEmails:")
	// TRACE.Println(jsondata)
	// get user for correct email username
	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	if storeduser.Username == "" {
		return ""
	}
	var rows *sql.Rows
	var err error
	retEmails := make([]EmailRowStruct, 0)
	PostgresEmailstring := `"` + storeduser.Username + `@pinged.email"`
	// TRACE.Println("PostgresCIDsstrPostgresEmailstringing = " + string(PostgresEmailstring))
	if db.postgres_conn == nil {
		return "" // can't do anything with no database connection :(
	}
	if jsondata.M_time != "" {
		// use m_time to get most recent messages
		// TRACE.Println("Using m_time to get most recent messages from postgres")
		// need to put the timestamp in single quotes
		// VALID:
		// SELECT f_username, content, m_time FROM  "4730d9e6-b719-411a-b179-62f35528c7d5" WHERE m_time > '2015-06-12T19:16:29.119Z'::timestamptz
		rows, err = db.postgres_conn.Query(`SELECT from_email, to_emails, recv_email, subject, content, attachments, starred, unread, spam, draft, deleted, recv_time, m_time FROM ` + PostgresEmailstring + ` WHERE m_time > '` + jsondata.M_time + `'::timestamptz ORDER BY m_time DESC;`)
	} else {
		// no m_time, so get 20 most recent messages
		// TRACE.Println("getting most recent 20 messages from postgres")
		// Average ms for 20 rows: 497ms
		// Average ms for 50 rows: 469.25ms // I KNOW, RIGHT???
		PostgresString := `SELECT from_email, to_emails, recv_email, subject, content, attachments, starred, unread, spam, draft, deleted, recv_time, m_time FROM ` + PostgresEmailstring + ` ORDER BY m_time DESC;`
		// TRACE.Println("PostgresString = " + string(PostgresString))
		rows, err = db.postgres_conn.Query(PostgresString)
	}
	if err, ok := err.(*pq.Error); ok {
		ERROR.Println("pq error:", err.Code.Name())
		ERROR.Println(err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var SqlFromEmail string
			var SqlToEmails string
			var SqlRecvEmail string
			var SqlSubject string
			var SqlContent string
			var SqlAttachments string
			var SqlStarred bool
			var SqlUnread bool
			var SqlSpam bool
			var SqlDraft bool
			var SqlDeleted bool
			var SqlRecvTime time.Time
			var SqlM_time time.Time
			if err := rows.Scan(&SqlFromEmail, &SqlToEmails, &SqlRecvEmail, &SqlSubject, &SqlContent, &SqlAttachments, &SqlStarred, &SqlUnread, &SqlSpam, &SqlDraft, &SqlDeleted, &SqlRecvTime, &SqlM_time); err != nil {
				ERROR.Println(err)
			}
			rowStruct := EmailRowStruct{
				FromEmail:   SqlFromEmail,
				ToEmails:    SqlToEmails,
				RecvEmail:   SqlRecvEmail,
				Subject:     SqlSubject,
				Content:     SqlContent,
				Attachments: SqlAttachments,
				Starred:     SqlStarred,
				Unread:      SqlUnread,
				Spam:        SqlSpam,
				Draft:       SqlDraft,
				Deleted:     SqlDeleted,
				RecvTime:    SqlRecvTime,
				M_time:      SqlM_time,
			}
			// convoRow, err := ffjson.Marshal(rowStruct)
			if err != nil {
				ERROR.Println("Error in ffjson.Marshal(rowStruct)")
				ERROR.Println(err)
			}
			// TRACE.Println("convoRow: " + string(convoRow))
			retEmails = append(retEmails, rowStruct)
		}
	}

	retCmd := EmailDataStruct{
		Cmd:    "GetAllEmails",
		Emails: retEmails,
	}
	retCmdString, err := json.Marshal(retCmd)
	if err != nil {
		ERROR.Println("Error in ffjson.Marshal(retCmd)")
		ERROR.Println(err)
	}
	// TRACE.Println("GetConvoData returning " + string(retCmdString))
	return string(retCmdString)
}

// USER
func (db *Database) GetUserByUsername(data string) string {
	cmdJSON := CommonCmdStruct{}
	err := json.Unmarshal([]byte(data), &cmdJSON)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	user := db.GetAerospikeUser(strings.ToUpper(cmdJSON.Username))
	// remove some fields that we don't want to send back
	retUser := UserStruct{
		Username:      user.Username,
		UsernameUpper: user.UsernameUpper,
		Email:         user.Email,
		Phone:         user.Phone,
		PhoneGateway:  user.PhoneGateway,
		Friends:       user.Friends,
		ProfilePic:    user.ProfilePic,
	}
	// TRACE.Println("In GetUserByUsername, returning " + user.ToJSONStringWithCmd("GetUserByUsername"))
	return retUser.ToJSONStringWithCmd("GetUserByUsername")
}

func (db *Database) GetUserByEmail(data string) string {
	cmdJSON := CreateUserCmdStruct{} // it works, has both cmd and email
	err := json.Unmarshal([]byte(data), &cmdJSON)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	user := db.GetAerospikeUserByEmail(cmdJSON.Email)
	// TRACE.Println("In GetUserByUsername, returning " + user.ToJSONStringWithCmd("GetUserByUsername"))
	return user.ToJSONStringWithCmd("GetUserByEmail")
}

func (db *Database) MatchUsers(data string) string {
	cmdJSON := MatchUsersCmdStruct{} // it works, has both cmd and email
	err := json.Unmarshal([]byte(data), &cmdJSON)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	foundUsers := make([]MatchUsersReturnStruct, 0) // start off empty
	// loop through and add to foundUsers
	for _, phone := range cmdJSON.Phones {
		// JSON passed in is {'Name': contacts[i].displayName, 'PhoneNum': contacts[i].phoneNumbers[j].value}
		user := db.GetAerospikeUserByPhone(phone.PhoneNum)
		if user.Username != "" {
			newUser := MatchUsersReturnStruct{
				Username:    user.Username,
				DisplayName: phone.Name,
				ProfilePic:  user.ProfilePic,
				Phone:       user.Phone,
				Email:       user.Email,
			}
			foundUsers = append(foundUsers, newUser)
		}
	}
	for _, email := range cmdJSON.Emails {
		// JSON passed in is {'Name': contacts[i].displayName, 'Email': contacts[i].emails[j].value}
		user := db.GetAerospikeUserByEmail(email.Email)
		if user.Username != "" {
			newUser := MatchUsersReturnStruct{
				Username:    user.Username,
				DisplayName: email.Name,
				ProfilePic:  user.ProfilePic,
				Phone:       user.Phone,
				Email:       user.Email,
			}
			foundUsers = append(foundUsers, newUser)
		}
	}
	retUserString, err := ffjson.Marshal(foundUsers)
	if err != nil {
		ERROR.Println("Error in ffjson.Marshal(retUserString)")
		ERROR.Println(err)
	}
	return `{"cmd":"MatchUsers","MatchedUsers":` + string(retUserString) + `}`
}

// FRIENDS
func (db *Database) AddFriend(data string) string {
	jsondata := FriendCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	// get user, add friend, save back
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.UID))
	// make sure friend isn't already added to user
	friendFound := false
	userFriends := user.GetFriendStructs()
	for _, friend := range userFriends {
		if strings.ToUpper(friend.Username) == strings.ToUpper(jsondata.FriendUID) {
			friendFound = true
			break
		}
	}
	// check both incoming and outgoing friend requests
	userIncomingPendingFriends := user.GetIncomingPendingFriendStructs()
	for _, friend := range userIncomingPendingFriends {
		if strings.ToUpper(friend.Username) == strings.ToUpper(jsondata.FriendUID) {
			friendFound = true
			break
		}
	}

	userOutgoingFriendRequests := user.GetOutgoingPendingFriendStructs()
	for _, friend := range userOutgoingFriendRequests {
		if strings.ToUpper(friend.Username) == strings.ToUpper(jsondata.FriendUID) {
			friendFound = true
			break
		}
	}

	if !friendFound {
		TRACE.Println("FriendUID " + jsondata.FriendUID + " was not found as a friend, and is being added.")
		// get user, update PendingFriends, save back
		friend := db.GetAerospikeUser(strings.ToUpper(jsondata.FriendUID))
		// save outgoing friend request to user who requested it
		newOutgoingFriend := UserFriendStruct{
			Username:   friend.Username,
			ProfilePic: friend.ProfilePic,
		}
		userOutgoingFriendRequests = append(userOutgoingFriendRequests, newOutgoingFriend)
		user.SaveOutgoingPendingFriendStructs(userOutgoingFriendRequests)
		OutPendFriendBin := aerospike.NewBin("OutPendFriend", user.OutgoingPendingFriends)
		userkey, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.Username))
		if err == nil {
			TRACE.Println("Saving user.OutgoingPendingFriends to " + user.UsernameUpper)
			TRACE.Println("updated user.OutgoingPendingFriends = " + user.OutgoingPendingFriends)
			db.UpdateAerospikeSingleBin(userkey, OutPendFriendBin)
		}
		// add profile pic and user who requested the friend to be added
		newFriend := UserFriendStruct{
			Username:   user.Username,
			ProfilePic: user.ProfilePic,
			Message:    jsondata.Message,
		}
		// save friend
		friendPendingFriends := friend.GetIncomingPendingFriendStructs()
		friendPendingFriends = append(friendPendingFriends, newFriend)
		friend.SaveIncomingPendingFriendStructs(friendPendingFriends)
		InPendFriendBin := aerospike.NewBin("InPendFriend", friend.IncomingPendingFriends)
		friendkey, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(friend.Username))
		if err == nil {
			TRACE.Println("saving friend.IncomingPendingFriends to " + friend.UsernameUpper)
			TRACE.Println("updated friend.IncomingPendingFriends = " + friend.IncomingPendingFriends)
			db.UpdateAerospikeSingleBin(friendkey, InPendFriendBin)
		}

		// send to all active web devices for both users
		webstrFriend := `{"cmd":"AddIncomingFriend","Friend":{"Username":"` + user.Username + `","ProfilePic":"` + user.ProfilePic + `","Message":"` + jsondata.Message + `"}}`
		db.SendStringToWebDevices(friend.Web, webstrFriend)
		// now send to other friend
		webstrUser := `{"cmd":"AddOutgoingFriend","Friend":{"Username":"` + user.Username + `","ProfilePic":"` + user.ProfilePic + `","Message":"` + jsondata.Message + `"}}`
		db.SendStringToWebDevices(user.Web, webstrUser)
	}
	return "" // we did our sending above in if !friendFound
}

func (db *Database) AcceptFriendRequest(data string) string {
	jsondata := FriendCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	// get user, add friend, save back
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.UID))
	friend := db.GetAerospikeUser(strings.ToUpper(jsondata.FriendUID))
	// make sure friend isn't already added to user
	friendFound := false
	userFriends := user.GetFriendStructs()
	for _, friend := range userFriends {
		if strings.ToUpper(friend.Username) == strings.ToUpper(jsondata.FriendUID) {
			friendFound = true
			break
		}
	}

	if !friendFound {
		// TRACE.Println("FriendUID " + jsondata.FriendUID + " was not found as a friend, and is being added.")
		// remove from IncomingPendingFriends
		userIncomingPendingFriends := user.GetIncomingPendingFriendStructs()
		for i, pendingFriend := range userIncomingPendingFriends {
			if strings.ToUpper(pendingFriend.Username) == strings.ToUpper(jsondata.FriendUID) {
				userIncomingPendingFriends = append(userIncomingPendingFriends[:i], userIncomingPendingFriends[i+1:]...)
				user.SaveIncomingPendingFriendStructs(userIncomingPendingFriends)
				break
			}
		}
		// remove from friend's OutgoingPendingFriends
		friendOutgoingPendingFriends := friend.GetOutgoingPendingFriendStructs()
		for i, pendingFriend := range friendOutgoingPendingFriends {
			if strings.ToUpper(pendingFriend.Username) == strings.ToUpper(user.UsernameUpper) {
				friendOutgoingPendingFriends = append(friendOutgoingPendingFriends[:i], friendOutgoingPendingFriends[i+1:]...)
				friend.SaveOutgoingPendingFriendStructs(friendOutgoingPendingFriends)
				break
			}
		}
		// now add to friends
		newFriend := UserFriendStruct{
			Username:   friend.Username,
			ProfilePic: friend.ProfilePic,
		}
		userFriends = append(userFriends, newFriend)
		user.SaveFriendStructs(userFriends)
		// now save back
		db.SetAerospikeUser(user.UsernameUpper, user.ToAerospikeBins())
		// now append user to friend
		newUserFriend := UserFriendStruct{
			Username:   user.Username,
			ProfilePic: user.ProfilePic,
		}
		friendFriends := friend.GetFriendStructs()
		friendFriends = append(friendFriends, newUserFriend)
		friend.SaveFriendStructs(friendFriends)
		// save to aerospike
		db.SetAerospikeUser(friend.UsernameUpper, friend.ToAerospikeBins())

		// send to both friend and user on any active device
		webstrFriend := `{"cmd":"AcceptFriendRequest", "Friend":{"Username":"` + user.Username + `", "ProfilePic":"` + user.ProfilePic + `"}}`
		db.SendStringToWebDevices(friend.Web, webstrFriend)
		// now send to other friend
		webstrUser := `{"cmd":"AcceptFriendRequest", "Friend":{"Username":"` + friend.Username + `", "ProfilePic":"` + friend.ProfilePic + `"}}`
		db.SendStringToWebDevices(user.Web, webstrUser)
	}
	return ""
}

func (db *Database) DenyFriendRequest(data string) string {
	jsondata := FriendCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	// get user, deny friend, save back
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.UID))
	userPendingFriends := user.GetIncomingPendingFriendStructs()
	for i, pendingFriend := range userPendingFriends {
		if strings.ToUpper(pendingFriend.Username) == strings.ToUpper(jsondata.FriendUID) {
			userPendingFriends = append(userPendingFriends[:i], userPendingFriends[i+1:]...)
			user.SaveIncomingPendingFriendStructs(userPendingFriends)
			userbins := user.ToAerospikeBins()
			db.SetAerospikeUser(user.UsernameUpper, userbins)
			break
		}
	}

	// get friend, deny user, save back
	friend := db.GetAerospikeUser(strings.ToUpper(jsondata.FriendUID))
	friendOutgoingPendingFriends := friend.GetOutgoingPendingFriendStructs()
	TRACE.Println("friendOutgoingPendingFriends:")
	TRACE.Println(friendOutgoingPendingFriends)
	for i, pendingFriend := range friendOutgoingPendingFriends {
		TRACE.Println("comparing " + pendingFriend.Username + " == " + jsondata.UID)
		if strings.ToUpper(pendingFriend.Username) == strings.ToUpper(jsondata.UID) {
			friendOutgoingPendingFriends = append(friendOutgoingPendingFriends[:i], friendOutgoingPendingFriends[i+1:]...)
			friend.SaveOutgoingPendingFriendStructs(friendOutgoingPendingFriends)
			friendbins := friend.ToAerospikeBins()
			db.SetAerospikeUser(friend.UsernameUpper, friendbins)
			break
		}
	}

	// send to both friend and user on any active device
	webstrFriend := `{"cmd":"DenyFriendRequest", "Friend":"` + user.Username + `"}`
	db.SendStringToWebDevices(friend.Web, webstrFriend)
	// now send to other friend
	webstrUser := `{"cmd":"DenyFriendRequest", "Friend":"` + friend.Username + `"}`
	db.SendStringToWebDevices(user.Web, webstrUser)
	return "" // we sent to web devices above already
}

func (db *Database) RemoveFriend(data string) string {
	jsondata := FriendCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	// get user, remove friend, save back
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.UID))
	userFriends := user.GetFriendStructs()
	for i, pendingFriend := range userFriends {
		if strings.ToUpper(pendingFriend.Username) == strings.ToUpper(jsondata.FriendUID) {
			userFriends = append(userFriends[:i], userFriends[i+1:]...)
			user.SaveFriendStructs(userFriends)
			friendsBin := aerospike.NewBin("Friends", user.Friends)
			key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
			if err == nil {
				db.UpdateAerospikeSingleBin(key, friendsBin)
			}
			break
		}
	}

	// get friend, deny user, save back
	friend := db.GetAerospikeUser(strings.ToUpper(jsondata.FriendUID))
	friendFriends := friend.GetFriendStructs()
	TRACE.Println("friendFriends:")
	TRACE.Println(friendFriends)
	for i, pendingFriend := range friendFriends {
		TRACE.Println("comparing " + pendingFriend.Username + " == " + jsondata.UID)
		if strings.ToUpper(pendingFriend.Username) == strings.ToUpper(jsondata.UID) {
			friendFriends = append(friendFriends[:i], friendFriends[i+1:]...)
			friend.SaveFriendStructs(friendFriends)
			friendsBin := aerospike.NewBin("Friends", friend.Friends)
			key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(friend.UsernameUpper))
			if err == nil {
				db.UpdateAerospikeSingleBin(key, friendsBin)
			}
			break
		}
	}

	// send to both friend and user on any active device
	webstrFriend := `{"cmd":"RemoveFriend", "Friend":"` + user.Username + `"}`
	db.SendStringToWebDevices(friend.Web, webstrFriend)
	// now send to other friend
	webstrUser := `{"cmd":"RemoveFriend", "Friend":"` + friend.Username + `"}`
	db.SendStringToWebDevices(user.Web, webstrUser)
	return "" // we sent to web devices above already
}

func (db *Database) SaveAutoreplyMessage(data string) string {
	// get user
	jsondata := AutoreplyList{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	if user.AutoreplyMessage == jsondata.Message {
		return `{"cmd":"AddAutoreplyMessage","ret_msg":"Autoreply message has not changed."}` // autoreply message is the same, so don't do anything
	} else {
		// autoreply message is different, so set and clear for each CID
		user.AutoreplyMessage = jsondata.Message
		// now update user struct
		userCIDs := user.GetCIDStructs()
		for i, _ := range userCIDs {
			userCIDs[i].AutoreplySent = 0
		}
		user.SaveCIDStructs(userCIDs)
		// TRACE.Println("saving user: " + user.ToJSONString())
		bins := aerospike.BinMap{
			"CIDs":          user.CIDs,
			"AutoreplyNote": user.AutoreplyMessage,
		}
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
		if err == nil {
			db.WriteAerospikeMultipleBins(key, bins)
		} else { // return error
			return `{"cmd":"AddAutoreplyMessage","ret_msg":"An error occured trying to save the autoreply message on PingedChat servers."}`
		}
	}
	return `{"cmd":"AddAutoreplyMessage","ret_msg":"Autoreply message saved successfully on PingedChat servers."}`
}

func (db *Database) ChangeProfilePic(data string) string {
	// get user
	jsondata := UserStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}
	// get user, save profile pic back
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	user.ProfilePic = jsondata.ProfilePic
	profilePicBin := aerospike.NewBin("ProfilePic", user.ProfilePic)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, profilePicBin)
	}
	// loop through friends and update user's profile pic
	userFriends := user.GetFriendStructs()
	for _, element := range userFriends {
		friend := db.GetAerospikeUser(strings.ToUpper(element.Username))
		friendFriends := friend.GetFriendStructs()
		for i, e := range friendFriends {
			if strings.ToUpper(e.Username) == user.UsernameUpper {
				friendFriends[i].ProfilePic = user.ProfilePic
				friend.SaveFriendStructs(friendFriends)
				// save to aerospike
				friendsBin := aerospike.NewBin("Friends", friend.Friends)
				key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(friend.UsernameUpper))
				if err == nil {
					db.UpdateAerospikeSingleBin(key, friendsBin)
				}
				break // continue on to outer for loop
			}
		}
	}
	return user.ToJSONStringWithCmd("ChangeProfilePic")
}

func (db *Database) ChangeUserPhone(data string) string {
	jsondata := ChangeAccountStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	if storeduser.Username == "" {
		return ""
	}
	phonenum := validatePhoneAndGetPhoneGateway(jsondata.Phone)
	if phonenum == "" {
		ERROR.Printf("validatePhoneAndGetPhoneGateway in ChangeUserPhone says the number isn't any good!")
		return ""
	}
	// valid phone number and user if we didn't return above
	storeduser.Phone = formatPhoneNumber(jsondata.Phone)
	storeduser.PhoneGateway = phonenum
	phoneBin := aerospike.NewBin("Phone", storeduser.Phone)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, phoneBin)
	}
	phoneGatewayBin := aerospike.NewBin("PhoneGateway", storeduser.PhoneGateway)
	phoneGatewayKey, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(phoneGatewayKey, phoneGatewayBin)
	}
	return storeduser.ToJSONStringWithCmd("ChangeUserPhone")
}

func (db *Database) ChangeUserEmail(data string) string {
	jsondata := ChangeAccountStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error:", err)
		return ""
	}

	storeduser := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	if storeduser.Username == "" {
		return ""
	}
	// make sure email isn't registered to anybody else
	// TODO:  validation?
	if jsondata.Email == "" {
		return "" // can't have an empty email!
	}
	existingUser := db.GetAerospikeUserByEmail(jsondata.Email)
	if existingUser.Username != "" {
		return "" // we had a user already with that username
	}
	storeduser.Email = jsondata.Email
	emailBin := aerospike.NewBin("Email", storeduser.Email)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(storeduser.UsernameUpper))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, emailBin)
	}
	return storeduser.ToJSONStringWithCmd("ChangeUserEmail")
}

func (db *Database) GetS3PolicyData(data string) string {
	// TRACE.Println("in GetS3PolicyData")
	s3data := s3Sign()
	// TRACE.Println("s3Data = " + s3data)
	return s3data
}

// DA BIG BOYS
func (db *Database) AddMessageToConvo(data MessageStruct) bool {
	PostgresCIDstring := `"` + data.CID + `"`
	CurString := "INSERT INTO " + PostgresCIDstring + " (f_username, content, m_time) VALUES ($1, $2, $3)" // conversation ID is the table name
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		return false // can't do anything with no database connection :(
	}
	_, inserterr := db.postgres_conn.Exec(CurString, data.FromUsername, data.Content, data.M_time)
	if inserterr != nil {
		ERROR.Println("error:", inserterr)
		return false
	}
	return true
}

func (db *Database) HandleBots(message string) string {
	if strings.HasPrefix(strings.ToLower(message), giphyPrefix) {
		// handle giphy
		TRACE.Println("handling giphy bot")
		message = rockGiphy(message[len(giphyPrefix):])
	} else if strings.HasPrefix(strings.ToLower(message), forecastPrefix) {
		message = GetWeather(message[len(forecastPrefix):])
	}
	return message
}

func (db *Database) SendMessage(data string) string {
	jsondata := MessageStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in SendMessage Unmarshalling into MessageStruct:", err)
		return ""
	}
	message := db.HandleBots(jsondata.Content)
	jsondata.Content = message
	// add message to convo
	messageAdded := db.AddMessageToConvo(jsondata)
	if !messageAdded {
		ERROR.Println("Error in adding message to postgres")
	}

	// save any autoreplies to be sent
	var autoreplies []AutoreplyList

	for _, member := range jsondata.ToUIDs {
		// TRACE.Println("SendMessage member = " + string(member))
		recipient := db.GetAerospikeUser(member)
		if recipient.UsernameUpper == "" {
			continue // user not found
		}

		TRACE.Println("recipient: " + recipient.ToJSONString())

		// don't send push message to the user who sent the message.
		// otherwise they're sitting at their computer and getting pushes about it
		// or pushed to the very device they used to send the message!
		isSelf := (recipient.UsernameUpper == strings.ToUpper(jsondata.FromUsername))
		TRACE.Printf("isSelf: %d", isSelf)

		// do the real magic, sending to devices!
		hasAndroid := false
		hasFireos := false
		hasIos := false
		hasWeb := false

		if !isSelf {
			// android first
			TRACE.Println("len(recipient.Android) = ")
			TRACE.Println(len(recipient.Android))
			if len(recipient.Android) > 0 {
				// TRACE.Println("pushing to Android device")
				PushToAndroid(recipient.Android, jsondata)
				hasAndroid = true
			}

			// fireos next
			TRACE.Println("len(recipient.Fireos) = ")
			TRACE.Println(len(recipient.Fireos))
			if len(recipient.Fireos) > 0 {
				// TRACE.Println("pushing to Fireos device")
				PushToFireos(recipient.Fireos, jsondata)
				hasFireos = true
			}

			TRACE.Println("len(recipient.Ios) = ")
			TRACE.Println(len(recipient.Ios))
			if len(recipient.Ios) > 0 {
				// TRACE.Println("pushing to iOS device")
				PushToIos(recipient.Ios, jsondata)
				hasIos = true
			}
		} // if (!isSelf)

		webzString, err := ffjson.Marshal(jsondata) // do this for HandleBots content that's been updated
		if err != nil {
			ERROR.Println("Error in ffjson.Marshal(jsondata) in SendMessage")
		} else {
			// TRACE.Println("webzString = " + string(webzString))
			for _, webz := range recipient.Web {
				// TRACE.Println("web device: " + webz)
				if db.nats_encodedconn != nil {
					db.nats_encodedconn.Publish(webz, string(webzString))
				}
				hasWeb = true
			}
		}

		// send SMS if nothing else registered
		if !hasAndroid && !hasFireos && !hasIos && !hasWeb && !isSelf {
			// TODO: sendSMS
			TRACE.Println("sending SMS")
			if (strings.ToUpper(recipient.Username) != strings.ToUpper(jsondata.FromUsername)) && (recipient.Phone != "") {
				convoName := db.GetAerospikeConvoName(jsondata.CID)
				PushToSMS(recipient.PhoneGateway, jsondata, convoName, recipient.Username)
			}
		}

		// loop through user CIDs and modify m_time
		//        alternative is to just use convos->[CID]->m_time, but how does that get updated to the user?
		// also unread_count.  Shoot, may have to loop through anyways.
		recipientCIDs := recipient.GetCIDStructs()
		for index, element := range recipientCIDs {
			if element.CID == jsondata.CID {
				// TRACE.Println("Updating user CID at index ")
				// TRACE.Println(index)
				// TRACE.Println(" , element = ")
				// TRACE.Println(element)
				// check for autoreply
				if recipient.AutoreplyMessage != "" && element.AutoreplySent == 0 {
					TRACE.Println(recipient.Username + " needs to send an autoreply")
					autoreplies = append(autoreplies, AutoreplyList{recipient.Username, recipient.AutoreplyMessage})
					element.AutoreplySent = 1 // we're sending later, so mark it as such
				}
				element.M_time = jsondata.M_time
				element.UnreadCount++
				recipientCIDs[index] = element
				recipient.SaveCIDStructs(recipientCIDs)
				// TRACE.Println("saving recipient: " + recipient.ToJSONString())
				CIDsBin := aerospike.NewBin("CIDs", recipient.CIDs)
				key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(recipient.UsernameUpper))
				if err == nil {
					db.UpdateAerospikeSingleBin(key, CIDsBin)
				}
				break
			}
		}
	} // end for ToUIDs loop

	// update time of convo
	db.SetAerospikeConvoMtime(jsondata.CID, jsondata.M_time)

	// send autoreplies, only to web devices though (not worth a push notification)
	if len(autoreplies) > 0 {
		// loop through users and send message
		for _, reply := range autoreplies {
			// format message for sending
			// we'll use same jsondata as before, since it's going to the same group as before
			jsondata.Content = "AUTOREPLY: " + reply.Message
			jsondata.FromUsername = reply.Username
			autoReplyString, err := ffjson.Marshal(jsondata) // do this for HandleBots content that's been updated
			if err != nil {
				ERROR.Println("Error in ffjson.Marshal(jsondata) in SendMessage autoreplies")
			} else {
				// add to db
				messageAdded := db.AddMessageToConvo(jsondata)
				if !messageAdded {
					ERROR.Println("Error in adding message to postgres")
				}
				// send to users
				for _, member := range jsondata.ToUIDs {
					// we have a member who needs the autoreply sent to them
					recipient := db.GetAerospikeUser(member)
					if recipient.UsernameUpper == "" {
						continue // user not found
					}
					db.SendStringToWebDevices(recipient.Web, string(autoReplyString))
				}
			}
		}
	}

	return ""
}

func (db *Database) UpdateUserStatus(data string) string {
	jsondata := CmdConvoMember{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in UpdateUserStatus Unmarshalling into MessageStruct:", err)
		return ""
	}
	// TRACE.Println("jsondata.ReadTime: " + jsondata.ReadTime)
	// TRACE.Println(jsondata)
	convoMembersStrings := db.GetAerospikeConvoMembers(jsondata.CID)
	convoMembers := ToConvoMemberArray(convoMembersStrings)
	// TRACE.Println("convoMembers: ")
	// TRACE.Println(convoMembers)

	// update user
	UsernameUpper := strings.ToUpper(jsondata.Username)
	for i, e := range convoMembers {
		if strings.ToUpper(e.Username) == UsernameUpper {
			if jsondata.NewReadTime != "" {
				if convoMembers[i].ReadTime < jsondata.NewReadTime {
					convoMembers[i].ReadTime = jsondata.NewReadTime
				}
			}
			convoMembers[i].Typing = jsondata.Typing // possible bug if user is updating read_time while actively typing
			// cause then the code will default to False for JSON Unmarshal and set Active to False
			break
		}
	}
	// save back
	db.SetAerospikeConvoMembers(jsondata.CID, convoMembers.ToStringArray())
	// setting updated M_time to now so users know to get updated data when offline
	newMTime := getCurrentUTCISOTimeString()
	db.SetAerospikeConvoMtime(jsondata.CID, newMTime)
	// now loop through users and send update status string to them
	jsondata.Cmd = "UpdateUserStatus" // want to send Cmd back
	jsonBytes, err := ffjson.Marshal(jsondata)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal in UpdateUserStatus:")
		ERROR.Println(err)
	} else {
		jsonString := string(jsonBytes)
		if db.nats_encodedconn != nil {
			// TRACE.Println("looping through each member and publishing user status update to them")
			for _, e := range convoMembers {
				recipient := db.GetAerospikeUser(strings.ToUpper(e.Username))
				db.SendStringToWebDevices(recipient.Web, jsonString)
			}
		}
	}

	return ""
}

func (db *Database) UpdateUnreadCount(data string) string {
	jsondata := CmdConvoUnreadCount{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in UpdateUnreadCount Unmarshalling into MessageStruct:", err)
		return ""
	}
	// get user, update unread count, save back
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	CIDs := user.GetCIDStructs()
	for i, e := range CIDs {
		if e.CID == jsondata.CID {
			CIDs[i].UnreadCount = jsondata.UnreadCount
			user.SaveCIDStructs(CIDs)
			CIDsBin := aerospike.NewBin("CIDs", user.CIDs)
			key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
			if err == nil {
				db.UpdateAerospikeSingleBin(key, CIDsBin)
			}
			break // no need to continue on
		}
	}
	return ""
}

func (db *Database) UpdateConvoMtime(data string) string {
	jsondata := CmdConvoMtimeCount{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in UpdateConvoMtime Unmarshalling into MessageStruct:", err)
		return ""
	}
	// get user, update unread count, save back
	user := db.GetAerospikeUser(strings.ToUpper(jsondata.Username))
	CIDs := user.GetCIDStructs()
	for i, e := range CIDs {
		if e.CID == jsondata.CID {
			CIDs[i].M_time = jsondata.Mtime
			user.SaveCIDStructs(CIDs)
			CIDsBin := aerospike.NewBin("CIDs", user.CIDs)
			key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
			if err == nil {
				db.UpdateAerospikeSingleBin(key, CIDsBin)
			}
			break // no need to continue on
		}
	}
	return ""
}

func (db *Database) SendEmailInvite(data string) string {
	jsondata := InviteEmailStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in UpdateUserStatus Unmarshalling into MessageStruct:", err)
		return ""
	}
	// now split emails into recipients
	// emails should be delimited by ',' (a comma)
	recipients := strings.Split(jsondata.Emails, ",")
	// send email
	mandrill.Key = MANDRILL_API_KEY
	// you can test your API key with Ping
	err = mandrill.Ping()
	// everything is OK if err is nil
	if err != nil {
		ERROR.Println("error in SendEmailInvite mandrill.Ping()")
		ERROR.Println(err)
	}
	// for found users
	var foundUsers []FoundUserStruct
	// now loop through recipients and send async email invites
	for _, rec := range recipients {
		TRACE.Println("SendEmailInvite, rec = " + rec)
		// first look if email is registered to a user
		emailUser := db.GetAerospikeUserByEmail(rec)
		if emailUser.Username != "" {
			// we have a user!
			foundUsers = append(foundUsers, FoundUserStruct{
				Username:   emailUser.Username,
				ProfilePic: emailUser.ProfilePic,
				Email:      emailUser.Email,
			})
			continue // don't send an email invite to a user who already exists!
		}

		msg := mandrill.NewMessageTo(rec, "")
		msg.HTML = "Please visit http://www.pingedchat.com to create an account and start pinging with " + jsondata.FromUsername + "!"
		msg.Subject = jsondata.FromUsername + " wants you to join PingedChat!"
		msg.FromEmail = "noreply@pingedchat.com"
		msg.FromName = "PingedChat Inviter"
		res, err := msg.Send(false)
		if err != nil {
			ERROR.Println("error in SendEmailInvite msg.Send(true)")
			ERROR.Println(err)
			if res[0].Status != "sent" {
				ERROR.Println("res.Status in SendEmailInvite = " + res[0].Status)
				ERROR.Println("res.RejectionReason in SendEmailInvite = " + res[0].RejectionReason)
			}
		}
	}

	// now handle phone numbers
	// must verify correct phone number and send email/text
	phones := strings.Split(jsondata.Phones, ",")
	TRACE.Printf("phones : ", phones)
	for _, p := range phones {
		TRACE.Println("SendEmailInvite, p = " + p)
		// first see if phone number is registered to a user
		phoneUser := db.GetAerospikeUserByPhone(p)
		if phoneUser.Username != "" {
			// we have a user!
			foundUsers = append(foundUsers, FoundUserStruct{
				Username:   phoneUser.Username,
				ProfilePic: phoneUser.ProfilePic,
				Phone:      phoneUser.Phone,
			})
			continue // don't send an email invite to a user who already exists!
		}

		phone := validatePhoneAndGetPhoneGateway(p)
		TRACE.Println("SendEmailInvite, phone = " + phone)
		msg := mandrill.NewMessageTo(phone, phone)
		msg.HTML = "Please visit http://www.pingedchat.com to create an account and start pinging with " + jsondata.FromUsername + "!"
		msg.Subject = jsondata.FromUsername + " wants you to join PingedChat!"
		msg.FromEmail = "noreply@pingedchat.com"
		msg.FromName = "PingedChat Inviter"
		res, err := msg.Send(false)
		if err != nil {
			ERROR.Println("error in SendEmailInvite msg.Send(false)")
			ERROR.Println(err)
			if res[0].Status != "sent" {
				ERROR.Println("res.Status in SendEmailInvite = " + res[0].Status)
				ERROR.Println("res.RejectionReason in SendEmailInvite = " + res[0].RejectionReason)
			}
		}
	}

	// return found users if any!
	if len(foundUsers) > 0 {
		// send back found users
		userstr, err := ffjson.Marshal(foundUsers)
		if err != nil {
			ERROR.Println("error in ffjson.Marshal")
			ERROR.Println(err)
			return ""
		}
		return `{"cmd":"SendEmailInvite", "FoundUsers":` + string(userstr) + `}`
	} else {
		return ""
	}
}

// insert into postgres
func (db *Database) AddEmailToDb(Username string, From string, ToEmails string, RecvEmail string, Subject string, Content string, Attachments string, Starred bool, Unread bool, Spam bool, SentTime string, ModifiedTime string) bool {
	// email table structure:
	// [   FROM_EMAIL   |   TO_EMAILS    |  RECV_EMAIL  |  SUBJECT  |  CONTENT  | ATTACHMENTS |  STARRED  |  UNREAD  |   SPAM   |  DRAFT  |  DELETED  |  RECV_TIME  |    M_TIME   ]
	// [     varchar    |    varchar     |   varchar    |  varchar  |  varchar  |   varchar   |  boolean  |  boolean |  boolean | boolean |  boolean  | timestamptz |  timestamptz ]
	PostgresTableNamestring := `"` + Username + `@pinged.email` + `"`
	CurString := "INSERT INTO " + PostgresTableNamestring + " (from_email, to_emails, recv_email, subject, content, attachments, starred, unread, spam, draft, deleted, recv_time, m_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)" // conversation ID is the table name
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		var err error
		TRACE.Println("opening postgres connection in AddEmailToDb()")
		db.postgres_conn, err = sql.Open("postgres", "user=postgres password=postgres dbname=testdb host=pingedchat-postgrestest.ctnuy46b8e6b.us-east-1.rds.amazonaws.com")
		if err != nil {
			ERROR.Println("error opening postgres: ", err)
			return false
		}
	}
	if Attachments == "null" {
		// no attachments will send attachments as "null", we want an empty string though
		Attachments = ""
	}
	_, inserterr := db.postgres_conn.Exec(CurString, From, ToEmails, RecvEmail, Subject, Content, Attachments, Starred, Unread, Spam, false, false, SentTime, ModifiedTime) // unread by default, not in drafts or trash by default
	if inserterr != nil {
		ERROR.Println("error:", inserterr)
		return false
	}
	return true
}

func (db *Database) SendEmailMessage(data string) string {
	jsondata := SendEmailCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in SendEmailMessage Unmarshalling into MessageStruct:", err)
		return ""
	}
	// send email
	mandrill.Key = MANDRILL_API_KEY
	// you can test your API key with Ping
	err = mandrill.Ping()
	// everything is OK if err is nil
	if err != nil {
		ERROR.Println("error in SendEmailMessage mandrill.Ping()")
		ERROR.Println(err)
	}
	username := jsondata.FromEmail[:strings.Index(jsondata.FromEmail, "@pinged.email")]
	user := db.GetAerospikeUser(username)
	msg := mandrill.NewMessage()
	for _, rec := range jsondata.ToEmails {
		msg.AddRecipient(rec, "") // params are email, name
	}
	var mail_attachments []string
	// we should only save to S3 once, and that's when they're received
	for _, attachment := range jsondata.Attachments {
		sDec, _ := base64.StdEncoding.DecodeString(attachment.Binary)
		// TRACE.Println("attachment: ", string(sDec))
		msg.AddAttachment([]byte(sDec), attachment.FileName, attachment.FileType)
		fullS3FilePath := db.UploadToS3(sDec, attachment.FileName, user)
		mail_attachments = append(mail_attachments, fullS3FilePath)
	}
	msg.HTML = jsondata.Content
	msg.Subject = jsondata.Subject
	msg.FromEmail = jsondata.FromEmail
	msg.FromName = jsondata.FromEmail // TODO
	res, err := msg.Send(false)       // more than 10 recipients are sent as async anyways by Mandrill
	if err != nil {
		ERROR.Println("error in SendEmailMessage msg.Send(false)")
		ERROR.Println(err)
		if res[0].Status != "sent" {
			ERROR.Println("res.Status in SendEmailMessage = " + res[0].Status)
			ERROR.Println("res.RejectionReason in SendEmailMessage = " + res[0].RejectionReason)
		}
		return `{"cmd":"SendEmailMessage","Status":"` + res[0].Status + `"}`
	}
	// add email to database
	if jsondata.FromEmail == "" {
		return `{"cmd":"SendEmailMessage","Status":"FromEmail was invalid"}`
	} else if strings.Index(jsondata.FromEmail, "@pinged.email") < 0 {
		return `{"cmd":"SendEmailMessage","Status":"FromEmail was invalid"}`
	}
	ToEmailBytes, err := json.Marshal(jsondata.ToEmails)
	ToEmailStr := string(ToEmailBytes)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal in SendEmailMessage() marshalling ToEmails string")
		ERROR.Println(err)
	}
	t := time.Now().UTC()
	t_s := t.Format(ISO8601_SECONDS) // time string
	// RecvEmail will be empty string ("") because we are not receiving it, but sending it
	// make attachments a string
	AttachmentsBytes, err := json.Marshal(mail_attachments)
	AttachmentsStr := string(AttachmentsBytes)
	db.AddEmailToDb(username, jsondata.FromEmail, ToEmailStr, "", jsondata.Subject, jsondata.Content, AttachmentsStr, false, false, false, t_s, t_s)
	// update EmailMtime
	EmailMTimeBin := aerospike.NewBin("EmailMtime", t_s)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(username))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, EmailMTimeBin)
	}
	// format return string
	msgStr := `{"FromEmail":"` + jsondata.FromEmail + `",` +
		`"ToEmails":` + ToEmailStr + `,` +
		`"Subject":"` + jsondata.Subject + `",` +
		`"Content":"` + html.EscapeString(jsondata.Content) + `",` +
		`"Attachments":` + AttachmentsStr + `,` +
		`"RecvTime":"` + t_s + `",` +
		`"M_time":"` + t_s + `"` +
		`}`
	return `{"cmd":"SendEmailMessage","Status":"Sent","Msg":` + msgStr + `}`
}

func (db *Database) MarkEmailUnread(data string) string {
	jsondata := UpdateEmailCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in MarkEmailUnread Unmarshalling into MessageStruct:", err)
		return ""
	}
	if jsondata.EmailMtime == "" {
		t := time.Now().UTC()
		jsondata.EmailMtime = t.Format(ISO8601_SECONDS) // time string
	}
	PostgresTableName := `"` + jsondata.Username + `@pinged.email"`
	CurString := "UPDATE " + PostgresTableName + " SET unread = $1 , m_time = $2 WHERE from_email = $3 AND subject = $4 AND recv_time = $5"
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil in MarkEmailUnread") // can't do anything with no database connection :(
	}
	_, inserterr := db.postgres_conn.Exec(CurString, jsondata.Unread, jsondata.EmailMtime, jsondata.FromEmail, jsondata.Subject, jsondata.RecvTime)
	if inserterr != nil {
		ERROR.Println("error adding to Postgres table in MarkEmailUnread: ", inserterr)
		return `{"cmd":"MarkEmailUnread","Success":false}`
	}
	// update aerospike user struct EmailMtime
	EmailMTimeBin := aerospike.NewBin("EmailMtime", jsondata.EmailMtime)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(jsondata.Username))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, EmailMTimeBin)
	} else {
		return `{"cmd":"MarkEmailUnread","Success":false}`
	}
	return `{"cmd":"MarkEmailUnread","Success":true}`
}

func (db *Database) MarkEmailStarred(data string) string {
	jsondata := UpdateEmailCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in MarkEmailStarred Unmarshalling into MessageStruct:", err)
		return ""
	}
	if jsondata.EmailMtime == "" {
		t := time.Now().UTC()
		jsondata.EmailMtime = t.Format(ISO8601_SECONDS) // time string
	}
	PostgresTableName := `"` + jsondata.Username + `@pinged.email"`
	CurString := "UPDATE " + PostgresTableName + " SET starred = $1 , m_time = $2 WHERE from_email = $3 AND subject = $4 AND recv_time = $5"
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil in MarkEmailStarred") // can't do anything with no database connection :(
	}
	_, inserterr := db.postgres_conn.Exec(CurString, jsondata.Starred, jsondata.EmailMtime, jsondata.FromEmail, jsondata.Subject, jsondata.RecvTime)
	if inserterr != nil {
		ERROR.Println("error adding to Postgres table in MarkEmailStarred: ", inserterr)
		return `{"cmd":"MarkEmailStarred","Success":false}`
	}
	// update aerospike user struct EmailMtime
	EmailMTimeBin := aerospike.NewBin("EmailMtime", jsondata.EmailMtime)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(jsondata.Username))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, EmailMTimeBin)
	} else {
		return `{"cmd":"MarkEmailStarred","Success":false}`
	}
	return `{"cmd":"MarkEmailStarred","Success":true}`
}

// we'll use recv_time as the key for drafts
// each draft has it's own recv_time, which is when it's created.  we'll update m_time when it updates.
func (db *Database) AddNewDraft(data string) string {
	jsondata := UpdateEmailCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in MarkEmailDeleted Unmarshalling into MessageStruct:", err)
		return ""
	}
	if jsondata.EmailMtime == "" {
		t := time.Now().UTC()
		jsondata.EmailMtime = t.Format(ISO8601_SECONDS) // time string
	}
	PostgresTableName := `"` + jsondata.Username + `@pinged.email"`
	// first remove draft that may be there
	CurString := "DELETE FROM " + PostgresTableName + " WHERE draft = true AND recv_time = $1"
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil in AddNewDraft") // can't do anything with no database connection :(
	}
	_, deleteerr := db.postgres_conn.Exec(CurString, jsondata.RecvTime)
	if deleteerr != nil {
		ERROR.Println("error deleting from Postgres table in AddNewDraft: ", deleteerr)
	}
	// create ToEmails string
	ToEmailBytes, err := json.Marshal(jsondata.ToEmails)
	ToEmailStr := string(ToEmailBytes)
	// now re-insert
	CurString = "INSERT INTO " + PostgresTableName + " (from_email, to_emails, recv_email, subject, content, attachments, starred, unread, spam, draft, deleted, recv_time, m_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"
	// TRACE.Println("cur_string = " + CurString)
	// TODO: attachments.  Save?  Delete?
	_, inserterr := db.postgres_conn.Exec(CurString, jsondata.FromEmail, ToEmailStr, "", jsondata.Subject, jsondata.Content, "", false, false, false, true, false, jsondata.RecvTime, jsondata.EmailMtime)
	if inserterr != nil {
		ERROR.Println("error adding to Postgres table in AddNewDraft: ", inserterr)
		return `{"cmd":"AddNewDraft","Success":false}`
	}
	// update aerospike user struct EmailMtime
	EmailMTimeBin := aerospike.NewBin("EmailMtime", jsondata.EmailMtime)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(jsondata.Username))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, EmailMTimeBin)
	} else {
		return `{"cmd":"AddNewDraft","Success":false}`
	}
	return `{"cmd":"AddNewDraft","Success":true}`
}

func (db *Database) MarkEmailDeleted(data string) string {
	jsondata := UpdateEmailCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in MarkEmailDeleted Unmarshalling into MessageStruct:", err)
		return ""
	}
	if jsondata.EmailMtime == "" {
		t := time.Now().UTC()
		jsondata.EmailMtime = t.Format(ISO8601_SECONDS) // time string
	}
	PostgresTableName := `"` + jsondata.Username + `@pinged.email"`
	CurString := "UPDATE " + PostgresTableName + " SET deleted = $1 , m_time = $2 WHERE from_email = $3 AND subject = $4 AND recv_time = $5"
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil in MarkEmailDeleted") // can't do anything with no database connection :(
	}
	_, inserterr := db.postgres_conn.Exec(CurString, jsondata.Deleted, jsondata.EmailMtime, jsondata.FromEmail, jsondata.Subject, jsondata.RecvTime)
	if inserterr != nil {
		ERROR.Println("error adding to Postgres table in MarkEmailDeleted: ", inserterr)
		return `{"cmd":"MarkEmailDeleted","Success":false}`
	}
	// update aerospike user struct EmailMtime
	EmailMTimeBin := aerospike.NewBin("EmailMtime", jsondata.EmailMtime)
	key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(jsondata.Username))
	if err == nil {
		db.UpdateAerospikeSingleBin(key, EmailMTimeBin)
	} else {
		return `{"cmd":"MarkEmailDeleted","Success":false}`
	}
	return `{"cmd":"MarkEmailDeleted","Success":true}`
}

func (db *Database) RemoveDeletedEmails(data string) string {
	jsondata := UpdateEmailCmdStruct{}
	err := json.Unmarshal([]byte(data), &jsondata)
	if err != nil {
		ERROR.Println("error in RemoveDeletedEmails Unmarshalling into MessageStruct:", err)
		return ""
	}
	PostgresTableName := `"` + jsondata.Username + `@pinged.email"`
	CurString := "DELETE FROM " + PostgresTableName + " WHERE deleted = true"
	// TRACE.Println("cur_string = " + CurString)
	if db.postgres_conn == nil {
		ERROR.Println("db.postgres_conn == nil in RemoveDeletedEmails") // can't do anything with no database connection :(
	}
	_, deleteerr := db.postgres_conn.Exec(CurString)
	if deleteerr != nil {
		ERROR.Println("error deleting from Postgres table in RemoveDeletedEmails: ", deleteerr)
		return `{"cmd":"RemoveDeletedEmails","Success":false}`
	}
	return `{"cmd":"RemoveDeletedEmails","Success":true}`
}
