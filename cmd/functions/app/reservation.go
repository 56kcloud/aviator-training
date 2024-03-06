package main

import (
	"aviator/reservation"
	"aviator/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
)

// reservationCrud is a router to route API routes to the correct backend method
func reservationCrud(ctx context.Context, request events.APIGatewayProxyRequest, path string, stage string,
	reservationApi reservation.ReservationApiInterface, errorClient utils.ApiErrorClient) (events.APIGatewayProxyResponse, error) {
	reservationId := request.PathParameters["reservationId"]

	var responseBody []byte
	switch request.HTTPMethod {
	case http.MethodGet:
		switch path {
		case "/reservations":
			var input reservation.ListInput
			queryParams := request.QueryStringParameters
			limitString, ok := queryParams["limit"]
			if ok {
				var limit *int32
				i, err := strconv.ParseInt(limitString, 10, 64)
				limit = aws.Int32(int32(i))

				if err != nil {
					return errorClient.ClientError(400, errors.New("Invalid limit"))
				}
				input.Limit = limit
			}

			nextTokenStr, ok := queryParams["nextToken"]
			if ok {
				input.NextToken = &nextTokenStr
			}

			reservations, err := reservationApi.List(input)
			errorClient.SetLogger(reservationApi.Logger())
			if err != nil {
				return errorClient.AwsError(err)
			}
			responseBody, err = json.Marshal(reservations)
			if err != nil {
				return errorClient.AwsError(err)
			}
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(responseBody),
				Headers:    utils.ResponseHeaders(),
			}, nil
		case fmt.Sprintf("/reservations/%s", reservationId):
			reservation, err := reservationApi.Get(reservationId)
			errorClient.SetLogger(reservationApi.Logger())
			if err != nil {
				return errorClient.AwsError(err)
			}
			responseBody, err = json.Marshal(reservation)
			if err != nil {
				return errorClient.AwsError(err)
			}
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(responseBody),
			Headers:    utils.ResponseHeaders(),
		}, nil
	case http.MethodPost:
		switch path {
		case "/reservations":
			b := []byte(request.Body)
			var requestBody reservation.Reservation
			err := json.Unmarshal(b, &requestBody)
			if err != nil {
				return errorClient.ClientError(400, err)
			}

			reservation, err := reservationApi.CreateOrUpdate(requestBody)
			errorClient.SetLogger(reservationApi.Logger())
			if err != nil {
				return errorClient.AwsError(err)
			}

			responseBody, err = json.Marshal(reservation)
			if err != nil {
				return errorClient.ClientError(500, err)
			}

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusCreated,
				Body:       string(responseBody),
				Headers:    utils.ResponseHeaders(),
			}, nil
		}
	case http.MethodDelete:
		switch path {
		case fmt.Sprintf("/reservations/%s", reservationId):
			err := reservationApi.Delete(reservationId)
			if err != nil {
				return errorClient.AwsError(err)
			}
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
			Headers:    utils.ResponseHeaders(),
		}, nil
	case http.MethodPut:
		switch path {
		case fmt.Sprintf("/reservations/%s", reservationId):
			b := []byte(request.Body)
			var requestBody reservation.Reservation
			err := json.Unmarshal(b, &requestBody)
			if err != nil {
				return errorClient.ClientError(400, err)
			}
			requestBody.Id = reservationId

			reservation, err := reservationApi.CreateOrUpdate(requestBody)
			if err != nil {
				return errorClient.AwsError(err)
			}

			responseBody, err = json.Marshal(reservation)
			if err != nil {
				return errorClient.ClientError(500, err)
			}
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(responseBody),
				Headers:    utils.ResponseHeaders(),
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
			Headers:    utils.ResponseHeaders(),
			Body:       string(responseBody),
		}, nil
	}

	return errorClient.ClientError(400, errors.New("bad request"))
}
