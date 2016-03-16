package pcDatabase

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

const (
	AWS_SECRET_ACCESS_KEY = "FILL_ME"
	AWS_KEY               = "FILL_ME"
	policy_document       = `{"expiration": "2040-01-01T00:00:00Z",
      "conditions": [
        {"bucket": "FILL_ME"},
        ["starts-with", "$key", "uploads/"],
        {"acl": "public-read"},
        ["starts-with", "$Content-Type", ""],
      ]
    }`
)

// liberal inspiration from https://github.com/mitchellh/goamz/blob/master/s3/sign.go

func s3Sign() string {
	var b64 = base64.StdEncoding

	b64policy := b64.EncodeToString([]byte(policy_document))
	hash := hmac.New(sha1.New, []byte(AWS_SECRET_ACCESS_KEY))
	hash.Write([]byte(b64policy))
	signature := make([]byte, b64.EncodedLen(hash.Size()))
	b64.Encode(signature, hash.Sum(nil))

	return `{ "cmd":"GetS3PolicyData", "policy":"` + string(b64policy) + `","signature":"` + string(signature) + `","AWS_KEY":"` + string(AWS_KEY) + `"}`
}
