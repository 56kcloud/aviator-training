/*
Package utils provides utility methods.
*/
package utils

import (
	aviatorErrors "aviator/errors"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/smithy-go"
)

var ErrorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

type ErrorResponse struct {
	Message string `json:"message"`
}

func ResponseHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "*",
		"Access-Control-Allow-Headers": "*",
	}
}

type ApiErrorClient struct {
	Language string
	Logger   *slog.Logger
}

// Returns a new error client from the provided config.
func NewFromConfig(language string, logger *slog.Logger) *ApiErrorClient {
	return &ApiErrorClient{Language: language, Logger: logger}
}

func (c *ApiErrorClient) SetLogger(logger *slog.Logger) {
	c.Logger = logger
}

// Builds and returns an APIGatewayProxyResponse for when downstream AWS services return an error.
func (c *ApiErrorClient) AwsError(err error) (events.APIGatewayProxyResponse, error) {
	ErrorLogger.Println(err.Error())

	var apiErr smithy.APIError
	var aviatorError aviatorErrors.AviatorError
	var errorResponse ErrorResponse
	var responseBody []byte
	var statusCode int

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		default:
			statusCode = http.StatusBadRequest
			errorResponse = ErrorResponse{Message: fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), apiErr.ErrorMessage())}
			c.Logger.Warn("aws error", "code", statusCode, "message", apiErr.ErrorMessage())
		case "AccessDeniedException":
			statusCode = http.StatusUnauthorized
			errorResponse = ErrorResponse{Message: http.StatusText(http.StatusUnauthorized)}
			c.Logger.Warn("aws error", "code", statusCode)
		case "InternalErrorException":
			debug.PrintStack()
			c.Logger.Error("server error")
			statusCode = http.StatusInternalServerError
			errorResponse = ErrorResponse{Message: http.StatusText(http.StatusInternalServerError)}
		}

		responseBody, _ = json.Marshal(errorResponse)
	} else if errors.As(err, &aviatorError) {
		statusCode = aviatorError.ApiError
		if c.Language == "fr" {
			errorResponse = ErrorResponse{Message: aviatorError.Message.FR}
		} else {
			errorResponse = ErrorResponse{Message: aviatorError.Message.EN}
		}
		c.Logger.Info("aviator error", "id", aviatorError.Id, "code", statusCode, "message", aviatorError.Message.EN)
		responseBody, _ = json.Marshal(errorResponse)
	} else {
		debug.PrintStack()
		c.Logger.Error("server error")
		errorResponse = ErrorResponse{Message: http.StatusText(http.StatusInternalServerError)}
		responseBody, _ = json.Marshal(errorResponse)
		statusCode = http.StatusInternalServerError
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(responseBody),
		Headers:    ResponseHeaders(),
	}, nil
}

// Builds and returns an APIGatewayProxyResponse when an error occurs.
func (c *ApiErrorClient) ClientError(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	ErrorLogger.Println(err.Error())
	debug.PrintStack()
	c.Logger.Error("server error")

	errorResponse := ErrorResponse{Message: fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), err.Error())}
	responseBody, _ := json.Marshal(errorResponse)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(responseBody),
		Headers:    ResponseHeaders(),
	}, nil
}
