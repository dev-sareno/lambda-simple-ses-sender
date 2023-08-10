package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	//go get -u github.com/aws/aws-sdk-go
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	DefaultSender = "noreply@graphnetworks.com.au"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	DefaultRecipient = "info@graphnetworks.com.au"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	DefaultSubject = "Form Submitted - graphnetworks.com.au"

	// The character encoding for the email.
	CharSet = "UTF-8"
)

type Payload struct {
	Name            string `json:"name"`
	CompanyName     string `json:"companyName" binding:"required"`
	CompanyIndustry string `json:"companyIndustry" binding:"required"`
	EmailAddress    string `json:"emailAddress" binding:"required"`
	PhoneNumber     string `json:"phoneNumber" binding:"required"`
	Message         string `json:"message"`
}

func validateMethod(request events.APIGatewayV2HTTPRequest) error {
	method := request.RequestContext.HTTP.Method
	if method != "POST" {
		return fmt.Errorf("POST method is expected, got %s", method)
	}
	return nil
}

func validatePath(request events.APIGatewayV2HTTPRequest) error {
	if request.RawPath != "/submit" && request.RawPath != "/submit/" {
		return fmt.Errorf("/submit path is expected, got %s", request.RawPath)
	}
	return nil
}

func getPayload(request events.APIGatewayV2HTTPRequest) (Payload, error) {
	result := Payload{}
	body := request.Body

	// validate content type
	if request.Headers["content-type"] != "application/json" {
		return result, errors.New("invalid content type or body")
	}

	if request.IsBase64Encoded {
		// decode base64
		data, err := base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			return result, fmt.Errorf("unable to decode body. %s", err.Error())
		}
		body = string(data)
	}

	// decode json
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return result, fmt.Errorf("unable to unmarshall json. %s", err.Error())
	}

	return result, nil
}

func constructBody(payload Payload) string {
	body := `
	The following details have been submitted via <a href="https://graphnetworks.com.au/">https://graphnetworks.com.au/</a>
	<br/>
	<br/>
	Name: {name}
	<br/>
	Company Name: {companyName}
	<br/>
	Company Industry: {companyIndustry}
	<br/>
	Email Address: {emailAddress}
	<br/>
	Phone Number: {phoneNumber}
	<br/>
	Message: {message}
	`

	// check if set via env variable
	envEmailTemplate := os.Getenv("EMAIL_TEMPLATE")
	if envEmailTemplate != "" {
		body = envEmailTemplate
	}

	body = strings.ReplaceAll(body, "{name}", payload.Name)
	body = strings.ReplaceAll(body, "{companyName}", payload.CompanyName)
	body = strings.ReplaceAll(body, "{companyIndustry}", payload.CompanyIndustry)
	body = strings.ReplaceAll(body, "{emailAddress}", payload.EmailAddress)
	body = strings.ReplaceAll(body, "{phoneNumber}", payload.PhoneNumber)
	body = strings.ReplaceAll(body, "{message}", payload.Message)

	return body
}

func isAuthenticated(request events.APIGatewayV2HTTPRequest) bool {
	return os.Getenv("AUTHTOKEN") == request.Headers["x-authtoken"]
}

func sendEmail(content string) *events.APIGatewayProxyResponse {
	// read env variables
	recipient := DefaultRecipient
	envRecipient := os.Getenv("EMAIL_RECIPIENT")
	if envRecipient != "" {
		recipient = envRecipient
	}

	subject := DefaultSubject
	envSubject := os.Getenv("EMAIL_SUBJECT")
	if envSubject != "" {
		subject = envSubject
	}

	sender := DefaultSender
	envSender := os.Getenv("EMAIL_SENDER")
	if envSender != "" {
		sender = envSender
	}

	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2")},
	)

	if err != nil {
		log.Fatalf("an error has occurred while creating seesion. %s\n", err.Error())
	}

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(content),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return &events.APIGatewayProxyResponse{Body: fmt.Sprintf("unable to send email. %s", err.Error()), StatusCode: 400}
	}

	fmt.Println("Email Sent to address: " + recipient)
	fmt.Println(result)
	return nil
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Printf("ctx: %+v\n", ctx)
	fmt.Printf("request: %+v\n", request)

	if !isAuthenticated(request) {
		return &events.APIGatewayProxyResponse{Body: "unauthorized", StatusCode: 401}, nil
	}

	if err := validateMethod(request); err != nil {
		fmt.Println(err.Error())
		return &events.APIGatewayProxyResponse{Body: "page not found", StatusCode: 404}, nil
	}

	if err := validatePath(request); err != nil {
		fmt.Println(err.Error())
		return &events.APIGatewayProxyResponse{Body: "page not found", StatusCode: 404}, nil
	}

	payload, err := getPayload(request)
	if err != nil {
		fmt.Println(err.Error())
		return &events.APIGatewayProxyResponse{Body: "bad request", StatusCode: 400}, nil
	}

	htmlBody := constructBody(payload)

	if errorResponse := sendEmail(htmlBody); errorResponse != nil {
		return errorResponse, nil
	}

	return &events.APIGatewayProxyResponse{Body: "ok", StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}
