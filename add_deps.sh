#!/bin/sh
go get gopkg.in/igm/sockjs-go.v2/sockjs
go get github.com/aerospike/aerospike-client-go
go get github.com/lib/pq
go get github.com/apcera/nats
# go get github.com/alexjlockwood/gcm # use local for cordova fix
# go get github.com/anachronistic/apns
go get github.com/timehop/apns
go get github.com/pkar/hermes
go get github.com/twinj/uuid
go get golang.org/x/crypto/bcrypt
go get github.com/mlbright/forecast/v2
go get github.com/mostafah/mandrill
go get github.com/ronniekritou/gotelapi
go get github.com/pquerna/ffjson/ffjson

#aws
go get -u github.com/aws/aws-sdk-go/aws
go get -u github.com/aws/aws-sdk-go/aws/session
go get -u github.com/aws/aws-sdk-go/aws/credentials
go get -u github.com/aws/aws-sdk-go/service/s3

# for setting up secondary indexes on aerospike
#aql
#CREATE INDEX phoneindex ON users.username (Phone) STRING
#CREATE INDEX emailindex ON users.username (Email) STRING
