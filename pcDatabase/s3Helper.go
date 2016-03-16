package pcDatabase

import (
	"bytes"
	aerospike "github.com/aerospike/aerospike-client-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var (
	AWSConfig  *aws.Config
	AWSSession *session.Session
	AWSCreds   *credentials.Credentials
	s3Conn     *s3.S3
)

func NowISO8601() string {
	t := time.Now().UTC()
	return t.Format(ISO8601_SECONDS)
}

func ConfigureAWS() {
	// get credentials
	AWS_CREDENTIALS_FILEPATH, err := filepath.Abs("./aws_credentials")
	if err != nil {
		ERROR.Println("error in getting absolute filepath for AWS credentials: ", err)
	}
	if AWSCreds == nil {
		AWSCreds = credentials.NewSharedCredentials(AWS_CREDENTIALS_FILEPATH, "default")
		_, err := AWSCreds.Get()
		if err != nil {
			ERROR.Println("Error retreiving AWS credentials: ", err)
			AWSCreds = nil
		}
	}

	if AWSConfig == nil {
		AWSConfig = &aws.Config{
			Region:           aws.String("us-east-1"),
			Endpoint:         aws.String("https://FILL_ME.s3.amazonaws.com"), // FILL_ME
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      AWSCreds,
			LogLevel:         aws.LogLevel(aws.LogDebug), // LogOff, LogDebug
		}
	}

	if AWSSession == nil {
		AWSSession = session.New(AWSConfig)
	}

	if s3Conn == nil {
		// S3 service client the Upload manager will use.
		s3Conn = s3.New(AWSSession)
		// skip s3manager for right meow
		// Create an uploader with S3 client and default options
		// uploader := s3manager.NewUploaderWithClient(s3Svc)
	}
}

func CreateS3Params(bucketName string, path string, fileBytes *bytes.Reader, size int, fileType string) *s3.PutObjectInput {
	return &s3.PutObjectInput{
		Bucket: aws.String(bucketName), // required, this is apparently folder name in the bucket, probably because I pass bucket name in as endpoint
		Key:    aws.String(path),       // required
		ACL:    aws.String("public-read"),
		Body:   fileBytes,
		// ContentLength: aws.Long(size),
		ContentType: aws.String(fileType),
		Metadata: map[string]*string{
			"Key": aws.String("PingedChat"), //required
		},
		// see more at http://docs.aws.amazon.com/sdk-for-go/api/service/s3/S3.html#PutObject-instance_method
	}
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func (db *Database) UploadToS3(sDec []byte, filename string, user UserStruct) string {
	fileSize := math.Ceil(float64(len(sDec)) / float64(1024))
	fileBytes := bytes.NewReader(sDec) // convert to io.ReadSeeker type
	fileType := http.DetectContentType(sDec)

	// make sure AWS is configured
	ConfigureAWS()
	// create filename
	randFilename := RandomString(12)
	randFilename += "-" + filename
	s3FilePath := "uploads/email-attachments/" + user.UsernameUpper + "/" + randFilename
	params := CreateS3Params("", s3FilePath, fileBytes, len(sDec), fileType)
	// actually perform the upload
	resp, err := s3Conn.PutObject(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		ERROR.Println("S3 Put error:", err.Error())
		return ""
	} else {
		// add to user used quota
		TRACE.Println("adding to user.QuotaUsed in kb: ", uint32(fileSize))
		user.QuotaUsed += uint32(fileSize)
		// save back
		QuotaUsedBin := aerospike.NewBin("QuotaUsed", int(user.QuotaUsed))
		key, err := aerospike.NewKey(AEROSPIKE_USERS_NAMESPACE, AEROSPIKE_USERS_USERNAME_TABLE, strings.ToUpper(user.UsernameUpper))
		if err == nil {
			db.UpdateAerospikeSingleBin(key, QuotaUsedBin)
		}
	}

	// Pretty-print the response data.
	TRACE.Println("S3 response: ", resp)

	// add to attachments array
	fullS3FilePath := "https://pingedchat-us1.s3.amazonaws.com/" + s3FilePath
	return fullS3FilePath
}
