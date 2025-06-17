package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	message := fmt.Sprintf(
		"Hello %s!, AWS Lambda ops biblebrain-service is running!",
		request.QueryStringParameters["name"],
	)

	return events.APIGatewayProxyResponse{Body: message, StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}

// This is a simple AWS Lambda function that responds with a greeting message.
