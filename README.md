# lambda-simple-ses-sender
Simple AWS Lambda function for sending email using Simple Email Service (SES) written in Go 1.x

## Run
```
$ export GOARCH="amd64"   // adjust accordingly
$ export GOOS="linux"     // adjust accordingly
$ go build -o Handler main.go
$ zip lambda.zip Handler
```

## Lambda Configuration
Update your Lambda function and make that the Handler value is `Handler`.

## Lambda Deploy
Upload the Zip file direcly to your Lambda.

## Lambda Environment Variables
Email values such as sender is designed to be configurable using Environment Variable
- `AUTHTOKEN` - Static authentication token. Can be set as `X-AuthToken` HTTP Header in the client side.
- `EMAIL_RECIPIENT` - The override email recipient
- `EMAIL_SENDER` - The override email sender
- `EMAIL_SUBJECT` - The override subject of the email
- `EMAIL_TEMPLATE` - The override email template. Leave empty to use the default template. The allowed variables in the template are:
  - `{name}`
  - `{companyName}`
  - `{companyIndustry}`
  - `{emailAddress}`
  - `{phoneNumber}`
  - `{message}`

## Lambda Permissions
Correct permissions must be granted to Lambda function so that it can be executed, write CloudWatch Logs, and use SES to send email.

## IAM Role
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "lambda.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
```

## IAM Policy - Logs
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "logs:CreateLogGroup",
            "Resource": "arn:aws:logs:ap-southeast-2:111122223333:*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": [
                "arn:aws:logs:ap-southeast-2:111122223333:log-group:/aws/lambda/mycustom-email-sender:*"
            ]
        }
    ]
}
```

## IAM Policy - SES Send Emal
```json
{
    "Id": "ExampleAuthorizationPolicy",
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowSendEmailForExampleComIdentity",
            "Effect": "Allow",
            "Resource": "arn:aws:ses:ap-southeast-2:111122223333:identity/example.com",
            "Action": [
                "ses:SendEmail",
                "ses:SendRawEmail"
            ]
        }
    ]
}
```
