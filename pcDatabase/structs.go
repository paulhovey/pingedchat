package pcDatabase

import (
	// "encoding/json"
	aerospike "github.com/aerospike/aerospike-client-go"
	"github.com/pquerna/ffjson/ffjson"
	"strings"
	"time"
)

type UserStruct struct {
	Username               string
	UsernameUpper          string
	Password               string
	Email                  string
	Phone                  string
	PhoneGateway           string
	AutoreplyMessage       string
	Friends                string
	IncomingPendingFriends string
	OutgoingPendingFriends string
	CIDs                   string // JSON marshal
	ScheduledMessages      string // JSON marshal
	EmailMtime             string // holds a date when email was last updated
	Android                []string
	Fireos                 []string
	Ios                    []string
	Web                    []string
	ProfilePic             string
	Quota                  uint32
	QuotaUsed              uint32
	SecQuests              string // security questions
	// for json exporting
	Cmd string `json:"cmd,omitempty"`
}

func (user *UserStruct) ToJSONString() string {
	userstring, err := ffjson.Marshal(user)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
		return ""
	} else {
		return string(userstring)
	}
}

func (user *UserStruct) ToJSONStringWithCmd(cmd string) string {
	user.Cmd = cmd
	userstring, err := ffjson.Marshal(user)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
		return ""
	} else {
		return string(userstring)
	}
}

func (user *UserStruct) GetCIDStructs() []UserCIDStruct {
	var CIDs []UserCIDStruct
	if user.CIDs == "" {
		return make([]UserCIDStruct, 0)
	}
	err := ffjson.Unmarshal([]byte(user.CIDs), &CIDs)
	if err != nil {
		ERROR.Println("error in json.Unmarshal, user.CIDs = " + user.CIDs)
		ERROR.Println(err)
	}
	return CIDs
}

func (user *UserStruct) SaveCIDStructs(CIDs []UserCIDStruct) {
	CIDsByte, err := ffjson.Marshal(CIDs)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
	} else {
		// TRACE.Println("In SaveCIDStructs, CIDsByte = ")
		// TRACE.Println(string(CIDsByte))
		user.CIDs = string(CIDsByte)
	}
}

func (user *UserStruct) GetFriendStructs() []UserFriendStruct {
	var Friends []UserFriendStruct
	err := ffjson.Unmarshal([]byte(user.Friends), &Friends)
	if err != nil {
		ERROR.Println("error in json.Unmarshal, user.Friends = " + user.Friends)
		ERROR.Println(err)
		Friends = make([]UserFriendStruct, 0)
	}
	return Friends
}

func (user *UserStruct) SaveFriendStructs(Friends []UserFriendStruct) {
	FriendsByte, err := ffjson.Marshal(Friends)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
	} else {
		user.Friends = string(FriendsByte)
	}
}

func (user *UserStruct) GetIncomingPendingFriendStructs() []UserFriendStruct {
	var IncomingPendingFriends []UserFriendStruct
	err := ffjson.Unmarshal([]byte(user.IncomingPendingFriends), &IncomingPendingFriends)
	if err != nil {
		ERROR.Println("error in json.Unmarshal, user.IncomingPendingFriends = " + user.IncomingPendingFriends)
		ERROR.Println(err)
		IncomingPendingFriends = make([]UserFriendStruct, 0)
	}
	return IncomingPendingFriends
}

func (user *UserStruct) SaveIncomingPendingFriendStructs(IncomingPendingFriends []UserFriendStruct) {
	IncomingPendingFriendsByte, err := ffjson.Marshal(IncomingPendingFriends)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
	} else {
		user.IncomingPendingFriends = string(IncomingPendingFriendsByte)
	}
}

func (user *UserStruct) GetOutgoingPendingFriendStructs() []UserFriendStruct {
	var OutgoingPendingFriends []UserFriendStruct
	err := ffjson.Unmarshal([]byte(user.OutgoingPendingFriends), &OutgoingPendingFriends)
	if err != nil {
		ERROR.Println("error in json.Unmarshal, user.OutgoingPendingFriends = " + user.OutgoingPendingFriends)
		ERROR.Println(err)
		OutgoingPendingFriends = make([]UserFriendStruct, 0)
	}
	return OutgoingPendingFriends
}

func (user *UserStruct) SaveOutgoingPendingFriendStructs(OutgoingPendingFriends []UserFriendStruct) {
	OutgoingPendingFriendsByte, err := ffjson.Marshal(OutgoingPendingFriends)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
	} else {
		user.OutgoingPendingFriends = string(OutgoingPendingFriendsByte)
	}
}

func (user *UserStruct) GetSecurityQuestionStructs() []SecurityQuestionStruct {
	var SecurityQuestions []SecurityQuestionStruct
	err := ffjson.Unmarshal([]byte(user.SecQuests), &SecurityQuestions)
	if err != nil {
		ERROR.Println("error in json.Unmarshal, user.SecQuests = " + user.SecQuests)
		ERROR.Println(err)
		SecurityQuestions = make([]SecurityQuestionStruct, 0)
	}
	return SecurityQuestions
}

func (user *UserStruct) SaveSecurityQuestionStructs(SecurityQuestions []SecurityQuestionStruct) {
	SecurityQuestionsByte, err := ffjson.Marshal(SecurityQuestions)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
	} else {
		user.SecQuests = string(SecurityQuestionsByte)
	}
}

func (user *UserStruct) GetScheduledMessagesStructs() []ScheduledMessagesStruct {
	var ScheduledMessages []ScheduledMessagesStruct
	err := ffjson.Unmarshal([]byte(user.ScheduledMessages), &ScheduledMessages)
	if err != nil {
		ERROR.Println("error in json.Unmarshal, user.ScheduledMessages = ")
		ERROR.Println(user.ScheduledMessages)
		ERROR.Println(err)
		ScheduledMessages = make([]ScheduledMessagesStruct, 0)
	}
	return ScheduledMessages
}

func (user *UserStruct) SaveScheduledMessagesStructs(ScheduledMessages []ScheduledMessagesStruct) {
	ScheduledMessagesByte, err := ffjson.Marshal(ScheduledMessages)
	if err != nil {
		ERROR.Println("error in ffjson.Marshal")
		ERROR.Println(err)
	} else {
		user.ScheduledMessages = string(ScheduledMessagesByte)
	}
}

func (user *UserStruct) ToAerospikeBins() aerospike.BinMap {
	bins := aerospike.BinMap{
		"Username":                         user.Username,
		"UsernameUpper":                    user.UsernameUpper,
		"Password":                         user.Password,
		AEROSPIKE_USERS_USERNAME_EMAIL_BIN: user.Email,
		AEROSPIKE_USERS_USERNAME_PHONE_BIN: user.Phone,
		"PhoneGateway":                     user.PhoneGateway,
		"AutoreplyNote":                    user.AutoreplyMessage,
		"Friends":                          user.Friends,
		"InPendFriend":                     user.IncomingPendingFriends,
		"OutPendFriend":                    user.OutgoingPendingFriends,
		"CIDs":                             user.CIDs,
		"SchedMessages":                    user.ScheduledMessages,
		"EmailMtime":                       user.EmailMtime,
		"Android":                          user.Android,
		"Fireos":                           user.Fireos,
		"Ios":                              user.Ios,
		"Web":                              user.Web,
		"ProfilePic":                       user.ProfilePic,
		"Quota":                            int(user.Quota),
		"QuotaUsed":                        int(user.QuotaUsed),
		"SecQuests":                        user.SecQuests,
	}
	return bins
}

// helper function for FillUserWithAerospikeBins
func InterfaceArrayToStringArray(interfaceArr []interface{}) []string {
	stringArr := make([]string, len(interfaceArr))
	for index, element := range interfaceArr {
		stringArr[index] = element.(string)
	}
	return stringArr
}

func FillUserWithAerospikeBins(recbins aerospike.BinMap) UserStruct {
	defer func() {
		if r := recover(); r != nil {
			ERROR.Println("Recovered in FillUserWithAerspikeBins", r)
		}
	}()
	// TRACE.Println("in FillUserWithAerospikeBins, recbins = ")
	// TRACE.Println(recbins)
	user := UserStruct{}

	// about time to put in some error handling
	if _, ok := recbins["Username"]; ok {
		user.Username = recbins["Username"].(string)
	}
	if _, ok := recbins["UsernameUpper"]; ok {
		user.UsernameUpper = recbins["UsernameUpper"].(string)
	}
	if _, ok := recbins["Password"]; ok {
		user.Password = recbins["Password"].(string)
	}
	if _, ok := recbins[AEROSPIKE_USERS_USERNAME_EMAIL_BIN]; ok {
		user.Email = recbins[AEROSPIKE_USERS_USERNAME_EMAIL_BIN].(string)
	}
	if _, ok := recbins[AEROSPIKE_USERS_USERNAME_PHONE_BIN]; ok {
		user.Phone = recbins[AEROSPIKE_USERS_USERNAME_PHONE_BIN].(string)
	}
	if _, ok := recbins["PhoneGateway"]; ok {
		user.PhoneGateway = recbins["PhoneGateway"].(string)
	}
	if _, ok := recbins["AutoreplyNote"]; ok {
		user.AutoreplyMessage = recbins["AutoreplyNote"].(string)
	}
	if _, ok := recbins["CIDs"]; ok {
		user.CIDs = recbins["CIDs"].(string)
	}
	if _, ok := recbins["SchedMessages"]; ok {
		user.ScheduledMessages = recbins["SchedMessages"].(string)
	}
	if _, ok := recbins["EmailMtime"]; ok {
		user.EmailMtime = recbins["EmailMtime"].(string)
	}
	if _, ok := recbins["Friends"]; ok {
		user.Friends = recbins["Friends"].(string)
	}
	if _, ok := recbins["InPendFriend"]; ok {
		// TRACE.Println("ok reading recbins[InPendFriend]")
		user.IncomingPendingFriends = recbins["InPendFriend"].(string)
	}
	if _, ok := recbins["OutPendFriend"]; ok {
		// TRACE.Println("ok reading recbins[OutPendFriend]")
		user.OutgoingPendingFriends = recbins["OutPendFriend"].(string)
	}
	if _, ok := recbins["ProfilePic"]; ok {
		user.ProfilePic = recbins["ProfilePic"].(string)
	}
	if _, ok := recbins["Quota"]; ok {
		user.Quota = uint32(recbins["Quota"].(int)) // https://github.com/aerospike/aerospike-client-go/issues/62
	}
	if _, ok := recbins["QuotaUsed"]; ok {
		user.QuotaUsed = uint32(recbins["QuotaUsed"].(int)) // https://github.com/aerospike/aerospike-client-go/issues/62
	}
	if _, ok := recbins["SecQuests"]; ok {
		user.SecQuests = recbins["SecQuests"].(string)
	}
	if _, ok := recbins["Android"]; ok {
		user.Android = InterfaceArrayToStringArray(recbins["Android"].([]interface{}))
	}
	if _, ok := recbins["Fireos"]; ok {
		user.Fireos = InterfaceArrayToStringArray(recbins["Fireos"].([]interface{}))
	}
	if _, ok := recbins["Ios"]; ok {
		user.Ios = InterfaceArrayToStringArray(recbins["Ios"].([]interface{}))
	}
	if _, ok := recbins["Web"]; ok {
		user.Web = InterfaceArrayToStringArray(recbins["Web"].([]interface{}))
	}

	return user
}

type SecurityQuestionStruct struct {
	Question string // was going to be int, but string works too because it's part of 0.1 type array so can't use index
	Answer   string
}

type QuotaCmdStruct struct {
	Cmd       string `json:"cmd"`
	Username  string
	Quota     uint32
	QuotaUsed uint32
}

type ScheduledMessagesStruct struct {
	CID     string
	Time    string
	Content string
}

type ScheduledMessagesCmdStruct struct {
	Cmd      string `json:"cmd"`
	Username string
	CID      string
	Time     string
	Content  string
}

type PasswordResetUserStruct struct {
	Cmd       string `json:"cmd"`
	Username  string
	Questions []string
	Answers   []string `json:",omitempty"`
}

type ChangeUserPasswordStruct struct {
	Cmd         string `json:"cmd"`
	Username    string
	OldPassword string
	NewPassword string
}

type CommonCmdStruct struct {
	Cmd      string `json:"cmd"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Username string `json:"username"`
}

type CreateUserCmdStruct struct {
	Cmd       string `json:"cmd"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Token     string `json:"token"`
	Username  string `json:"username"`
	SecQuests string `json:"secQuests"`
}

type ValidateUserCmdStruct struct {
	Cmd      string `json:"cmd"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Username string `json:"username"`
}

type ChangeDeviceStruct struct {
	Cmd       string `json:"cmd"`
	Username  string `json:"username"`
	Device    string `json:"device"`
	OldDevice string `json:"old_device"`
}

type ChangeAccountStruct struct {
	Cmd      string `json:"cmd"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type MatchUsersCmdPhoneStruct struct {
	Name     string
	PhoneNum string
}

type MatchUsersCmdEmailStruct struct {
	Name  string
	Email string
}

type MatchUsersCmdStruct struct {
	Cmd    string `json:"cmd"`
	Phones []MatchUsersCmdPhoneStruct
	Emails []MatchUsersCmdEmailStruct
}

type MatchUsersReturnStruct struct {
	Username    string
	DisplayName string
	ProfilePic  string
	Phone       string
	Email       string
}

// convo structs
type CIDCommandStruct struct {
	Cmd      string   `json:"cmd,omitempty"`
	CID      string   `json:"CID"`
	Name     string   `json:"Name"`
	M_time   string   `json:"M_time"`
	Username string   `json:"Username,omitempty"`
	Members  []string `json:"Members"`
	Messages []string `json:"Messages,omitempty"`
}

type ConvoDataStruct struct {
	Cmd      string           `json:"cmd,omitempty"`
	CID      string           `json:"CID,omitempty"`
	Name     string           `json:"Name,omitempty"`
	M_time   string           `json:"M_time,omitempty"`
	Members  []string         `json:"Members,omitempty"`
	Files    []string         `json:"Files,omitempty"`
	Messages []ConvoRowStruct `json:"Messages,omitempty"`
}

type UserCIDStruct struct {
	CID           string
	M_time        string
	UnreadCount   int
	AutoreplySent int
}

type AutoreplyList struct {
	Username string
	Message  string
}

type UserFriendStruct struct {
	Username   string
	ProfilePic string
	Message    string
}

type ConvoRowStruct struct {
	// CID        string `json:"CID"`
	F_username string    `json:"f_username"`
	M_time     time.Time `json:"m_time"`
	Content    string    `json:"content"`
}

// convo members
type ConvoMember struct {
	Username string `json:"username"`
	ReadTime string `json:"read_time"`
	Typing   bool   `json:"typing"`
}

type ConvoMemberArray []ConvoMember

// sorting functions, tested at https://play.golang.org/p/fkYLc-DvBQ
// returns the number of elements in the collection.
func (slice ConvoMemberArray) Len() int {
	return len(slice)
}

// returns whether the element with index i should sort before
// the element with index j.
func (slice ConvoMemberArray) Less(i, j int) bool {
	return strings.ToUpper(slice[i].Username) < strings.ToUpper(slice[j].Username)
}

func (slice ConvoMemberArray) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (arr *ConvoMemberArray) ToStringArray() []string {
	var stringArray []string
	for _, element := range *arr {
		// TRACE.Println("ConvoMemberArray.ToStringArray() element = ")
		// TRACE.Println(element)
		str, err := ffjson.Marshal(element)
		if err != nil {
			ERROR.Println("error in ffjson.Marshal")
			ERROR.Println(err)
		} else {
			stringArray = append(stringArray, string(str))
		}
	}
	return stringArray
}

func ToConvoMemberArray(arr []string) ConvoMemberArray {
	var memberArray ConvoMemberArray
	var member ConvoMember

	for _, element := range arr {
		// TRACE.Println("StringArray.ToConvoMemberArray() element = ")
		// TRACE.Println(element)
		err := ffjson.Unmarshal([]byte(element), &member)
		if err != nil {
			ERROR.Println("error in json.Unmarshal")
			ERROR.Println(err)
		} else {
			memberArray = append(memberArray, member)
		}
	}
	return memberArray
}

type CmdConvoMember struct {
	Cmd         string `json:"cmd,omitempty"`
	CID         string
	Username    string
	OldReadTime string
	NewReadTime string
	Typing      bool
}

type CmdConvoUnreadCount struct {
	Cmd         string `json:"cmd,omitempty"`
	CID         string `json:"CID"`
	Username    string `json:"username"`
	UnreadCount int    `json:"unread_count"`
}

type CmdConvoMtimeCount struct {
	Cmd      string `json:"cmd,omitempty"`
	CID      string `json:"CID"`
	Username string `json:"Username"`
	Mtime    string `json:"M_time"`
}

func (arr *ConvoFileListArray) ToStringArray() []string {
	var stringArray []string
	for _, element := range *arr {
		// TRACE.Println("ConvoMemberArray.ToStringArray() element = ")
		// TRACE.Println(element)
		str, err := ffjson.Marshal(element)
		if err != nil {
			ERROR.Println("error in ffjson.Marshal")
			ERROR.Println(err)
		} else {
			stringArray = append(stringArray, string(str))
		}
	}
	return stringArray
}

func ToConvoFileListArray(arr []string) ConvoFileListArray {
	var fileArray ConvoFileListArray
	var file ConvoFileListStruct

	for _, element := range arr {
		// TRACE.Println("StringArray.ToConvoMemberArray() element = ")
		// TRACE.Println(element)
		err := ffjson.Unmarshal([]byte(element), &file)
		if err != nil {
			ERROR.Println("error in json.Unmarshal")
			ERROR.Println(err)
		} else {
			fileArray = append(fileArray, file)
		}
	}
	return fileArray
}

type CmdConvoAddFileStruct struct {
	Cmd          string `json:"cmd,omitempty"`
	CID          string `json:"CID"`
	FromUsername string `json:"f_username"`
	FileURL      string `json:"fileURL"`
	M_time       string `json:"m_time"`
}

type ConvoFileListArray []ConvoFileListStruct
type ConvoFileListStruct struct {
	FromUsername string `json:"f_username"`
	FileURL      string `json:"fileURL"`
	M_time       string `json:"m_time"`
}

// EMAILS
// email table structure:
// [   FROM_EMAIL   |   TO_EMAILS    |  RECV_EMAIL  |  SUBJECT  |  CONTENT  | ATTACHMENTS |  STARRED  |  UNREAD  |   SPAM   |  DRAFT  |  DELETED  |  RECV_TIME  |    M_TIME   ]
// [     varchar    |    varchar     |   varchar    |  varchar  |  varchar  |   varchar   |  boolean  |  boolean |  boolean | boolean |  boolean  | timestamptz |  timestamptz ]
type EmailRowStruct struct {
	FromEmail   string    `json:"from_email"`
	ToEmails    string    `json:"to_emails"`
	RecvEmail   string    `json:"recv_email"`
	Subject     string    `json:"subject"`
	Content     string    `json:"content"`
	Attachments string    `json:"attachments"`
	Starred     bool      `json:"starred"`
	Unread      bool      `json:"unread"`
	Spam        bool      `json:"spam"`
	Draft       bool      `json:"draft"`
	Deleted     bool      `json:"deleted"`
	RecvTime    time.Time `json:"recv_time"`
	M_time      time.Time `json:"m_time"`
}

type EmailDataStruct struct {
	Cmd    string           `json:"cmd,omitempty"`
	Emails []EmailRowStruct `json:"Messages,omitempty"`
}

type SendEmailCmdStruct struct {
	Cmd         string `json:"cmd"`
	FromEmail   string
	ToEmails    []string
	Subject     string
	Content     string
	Attachments []SendEmailAttachment
	M_time      string
}

type SendEmailAttachment struct {
	FileType string
	FileName string
	Binary   string
}

// used for updating fields in the database from the client
type UpdateEmailCmdStruct struct {
	Cmd         string `json:"cmd"`
	Username    string
	FromEmail   string
	ToEmails    []string
	Subject     string
	Attachments []SendEmailAttachment
	Content     string
	RecvTime    string
	EmailMtime  string
	Unread      bool
	Spam        bool
	Starred     bool
	Deleted     bool
}

// FRIENDS
type FriendCmdStruct struct {
	Cmd       string `json:"cmd"`
	UID       string `json:"UID"`
	FriendUID string `json:"friend_UID"`
	Message   string `json:"Message"`
}

// messages
type MessageStruct struct {
	Cmd          string   `json:"cmd,omitempty"`
	CID          string   `json:"CID,omitempty"`
	FromUsername string   `json:"f_username,omitempty"`
	ToUIDs       []string `json:"t_UIDs,omitempty"`
	M_time       string   `json:"m_time,omitempty"`
	Content      string   `json:"content,omitempty"`
}

// adm
type ADMResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// invite users
type InviteEmailStruct struct {
	Cmd          string `json:"cmd"`
	FromUsername string `json:"f_username"`
	Emails       string `json:"emails"`
	Phones       string `json:"phones"`
}

type FoundUserStruct struct {
	Username   string
	ProfilePic string
	Email      string
	Phone      string
}
