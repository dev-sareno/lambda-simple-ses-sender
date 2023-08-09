package main

import (
	"context"
	"fmt"
	"log"
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
	Sender = "noreply@graphnetworks.com.au"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	Recipient = "hpo.sareno@gmail.com"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	Subject = "Form Submitted - graphnetworks.com.au"

	// The character encoding for the email.
	CharSet = "UTF-8"
)

type Payload struct {
	Name            string
	CompanyName     string
	CompanyIndustry string
	EmailAddress    string
	PhoneNumber     string
	Message         string
}

func validateMethod(request events.APIGatewayProxyRequest) error {
	// TODO:
	return nil
}

func validatePath(request events.APIGatewayProxyRequest) error {
	// TODO:
	return nil
}

func getPayload(request events.APIGatewayProxyRequest) (Payload, error) {
	// TODO:
	return Payload{}, nil
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

	body += strings.ReplaceAll(body, "{name}", payload.Name)
	body += strings.ReplaceAll(body, "{companyName}", payload.CompanyName)
	body += strings.ReplaceAll(body, "{companyIndustry}", payload.CompanyIndustry)
	body += strings.ReplaceAll(body, "{emailAddress}", payload.EmailAddress)
	body += strings.ReplaceAll(body, "{phoneNumber}", payload.PhoneNumber)
	body += strings.ReplaceAll(body, "{message}", payload.Message)

	return body
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Printf("ctx: %+v\n", ctx)
	fmt.Printf("request: %+v\n", request)

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
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(htmlBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
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

		return &events.APIGatewayProxyResponse{Body: fmt.Sprintf("unable to send email. %s", err.Error()), StatusCode: 400}, nil
	}

	fmt.Println("Email Sent to address: " + Recipient)
	fmt.Println(result)

	return &events.APIGatewayProxyResponse{Body: "ok", StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}
