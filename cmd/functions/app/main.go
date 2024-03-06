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

func handle(ctx context.Context, request events.APIGatewayProxyRequest, reservationApi reservation.ReservationApiInterface) (events.APIGatewayProxyResponse, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("handling of api call started...")

	errorClient := utils.NewFromConfig("en", logger)

	stage := request.StageVariables["name"]
	path := strings.TrimPrefix(request.Path, "/"+stage)

	logger = logger.With("method", request.HTTPMethod)
	logger = logger.With("path", path)
	logger.Info("handling of api call started...")

	if strings.HasPrefix(path, "/reservations") {
		reservationApi.SetLogger(logger)
		return reservationCrud(ctx, request, path, stage, reservationApi, *errorClient)
	}

	return errorClient.ClientError(400, errors.New("bad request"))
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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

	return handle(ctx, request, reservationClient)

}

func main() {
	lambda.Start(HandleRequest)
}
