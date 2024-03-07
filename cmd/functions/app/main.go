/*
Lambda app handles all API calls.
*/
package main

import (
	"aviator/database"
	"aviator/reservation"
	"aviator/utils"
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	stage := request.StageVariables["name"]
	path := strings.TrimPrefix(request.Path, "/"+stage)

	logger = logger.With("method", request.HTTPMethod)
	logger = logger.With("path", path)
	logger.Info("handling of api call started...")

	conf, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		errorClient := utils.NewFromConfig("en", logger)
		return errorClient.AwsError(err)
	}

	databaseClient := database.NewFromConfig(
		database.Config{
			DynamoDbClient: dynamodb.NewFromConfig(conf),
			TableName:      os.Getenv("DYNAMODB_TABLE_NAME"),
		},
	)

	reservationClient := reservation.NewFromConfig(
		reservation.Config{
			Logger:         logger,
			DatabaseClient: *databaseClient,
		},
	)

	errorClient := utils.NewFromConfig("en", logger)

	if strings.HasPrefix(path, "/reservations") {
		reservationClient.SetLogger(logger)
		return reservationCrud(ctx, request, path, stage, reservationClient, *errorClient)
	}

	return errorClient.ClientError(400, errors.New("bad request"))
}

func main() {
	lambda.Start(HandleRequest)
}
