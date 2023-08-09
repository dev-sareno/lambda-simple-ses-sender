# lambda-simple-ses-sender
Simple AWS Lambda function for sending email using Simple Email Service (SES) written in Go 1.x

## Run
```
$ export GOARCH="amd64"   // adjust accordingly
$ export GOOS="linux"     // adjust accordingly
$ go build -o Handler main.go
$ zip lambda.zip Handler
```
