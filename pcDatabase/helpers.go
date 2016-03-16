package pcDatabase

import (
	"github.com/ronniekritou/gotelapi"
	"runtime/debug"
	"strings"
	"time"
)

const (
	ISO8601_SECONDS = "2006-01-02T15:04:05Z0700"
)

func Protect(g func()) {
	defer func() {
		if r := recover(); r != nil {
			ERROR.Printf("Runtime panic: %v\n", r)
			ERROR.Printf("Stack trace: %s\n", debug.Stack())
			err, ok := r.(error)
			if !ok {
				ERROR.Printf("recovered from err: ", err)
			}
			ERROR.Println("restarting function in Protect()")
			Protect(g)
		}
	}()
	TRACE.Println("start function in Protect()")
	g()
}

// ISO8601
func getCurrentUTCISOTimeString() string {
	t := time.Now().UTC()
	return t.Format(ISO8601_SECONDS)
}

var invalidPhoneNumbers = [...]string{
	"911",
	"+1911",
}
var telapi_helper telapi.TelapiHelper

func formatPhoneNumber(input string) string {
	// will return empty string if not valid phone number
	// valid phone number starts with country code, "+1..."
	if len(input) == 0 {
		return ""
	}
	if string(input[0]) != "+" {
		// we do not have our country code in place
		return ""
	}
	// check against invalid phone numbers
	for _, invalid := range invalidPhoneNumbers {
		if input == invalid {
			return "" // can't have an invalid number!  return empty.
		}
	}

	input = strings.TrimSpace(input)            // trim any excess blank characters
	input = strings.Replace(input, "-", "", -1) // remove all '-' characters
	input = strings.Replace(input, ".", "", -1) // remove all '.' characters
	input = strings.Replace(input, "(", "", -1) // remove all '(' characters
	input = strings.Replace(input, ")", "", -1) // remove all ')' characters
	// at this point, we should have a string that looks like +10123456789
	if len(input) < 12 || len(input) > 13 {
		// country codes can be one or two digits, so valid phone numbers can only be 12 or 13 digits including '+' character
		return ""
	}
	return input
}

func formatPhoneSMSEmailGateway(phonenum string, carrier string) string {
	// hardcoded lookup for TelApi carrier and format email gateway
	// good reference:  http://martinfitzpatrick.name/list-of-email-to-sms-gateways/
	switch carrier {
	case "T-Mobile USA, Inc.":
		return phonenum + "@tmomail.net"
	case "Verizon Wireless":
		return phonenum + "@vtext.com"
	case "Sprint Spectrum, L.P.":
		return phonenum + "@messaging.sprintpcs.com"
	case "AT&T Mobility":
		return phonenum + "@txt.att.net"
	case "Virgin Mobile - Sprint Reseller":
		return phonenum + "@vmobl.com"
	}
	return ""
}

func validatePhoneAndGetPhoneGateway(phone string) string {
	// I hate putting "and" in the title but I want this to do two things, soooo
	// returns either a blank string meaning no valid phone number and gateway
	// or returns a valid email gateway for SMS
	phonenum := formatPhoneNumber(phone)
	TRACE.Println("validatePhoneAndGetPhoneGateway, phone = " + phone + " phonenum = " + phonenum)
	if phonenum == "" {
		// invalid phone number
		return ""
	}
	// lookup user phone number
	if telapi_helper.AuthToken == "" {
		var err error
		telapi_helper, err = telapi.CreateClient(TELAPI_TESTING_SID, TELAPI_TESTING_AUTH)
		if err != nil {
			ERROR.Printf("telapi.CreateClient err: ", err)
		}
	}
	carrier, err := telapi_helper.CarrierLookup(phonenum)
	if err != nil {
		ERROR.Printf("telapi.CarrierLookup err: ", err)
	}
	if carrier.Mobile != true {
		ERROR.Println("phone number is not mobile number")
		return ""
	}
	TRACE.Println("carrier.Network = " + carrier.Network)
	return formatPhoneSMSEmailGateway(phonenum, carrier.Network)
}
