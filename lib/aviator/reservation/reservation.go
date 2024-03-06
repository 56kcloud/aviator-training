/*
Package reservation provides methods for performing CRUD operations on reservations.
*/
package reservation

import (
	"aviator/database"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/oklog/ulid/v2"
)

const CLUB_PARTITION_KEY = "CLUB"

// In a real app, CLUB_ID would be a dynamic variable
const CLUB_ID = "01HR9ZZNRFCKMAYNW3RY561QCP"
const RESERVATION_PARTITION_KEY = "RESERVATION"

type ReservationApiInterface interface {
	Logger() *slog.Logger
	SetLogger(logger *slog.Logger)
	CreateOrUpdate(input Reservation) (*Reservation, error)
	Get(reservationId string) (*Reservation, error)
	List(input ListInput) (*ListOutput, error)
	Delete(reservationId string) error
}

type Config struct {
	Logger         *slog.Logger
	DatabaseClient database.Client
	TenantId       string
	UserId         string
	UserRole       string
}

type Client struct {
	Config
}

// Item used to store a reservation
type Reservation struct {
	// Reservation Id: e.g. 01H55420KY47HRVVPK1Z3BSACK
	Id string `json:"id"`
	// Reserved aircraft Id: e.g. HB-KFQ
	Aircraft string `json:"aircraft"`
	// Flight type
	ReservationType string `json:"reservationType"`
	// Pilot who will fly, usually the same as Booker
	Pilot string `json:"pilot"`
	// Instructor (if any)
	Instructor *string `dynamodbav:",omitempty" json:"instructor"`
	// Start time of the reservation
	StartTime time.Time `json:"startTime"`
	// End time of the reservation
	EndTime time.Time `json:"endTime"`
	// Any remarks the booker wants to set for this reservation: e.g. "Short flight to the Matterhorn"
	Remarks   string    `json:"remarks"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Database item to store the reservation.
type databaseItem struct {
	// Primary key: e.g. RESERVATION#01GYEVQ6JTB0VZDJYHSEKTA46R
	PK string
	// Sort key: e.g. RESERVATION#01H55420KY47HRVVPK1Z3BSACK
	SK string

	// Item type: reservation
	ItemType string
	Reservation
}

// Returns a new reservation API client from the provided config.
func NewFromConfig(c Config) *Client {
	return &Client{Config: c}
}

func (c *Client) Logger() *slog.Logger {
	return c.Config.Logger
}

func (c *Client) SetLogger(logger *slog.Logger) {
	c.Config.Logger = logger
}

// Create or update a reservation
func (c *Client) CreateOrUpdate(input Reservation) (*Reservation, error) {
	newReservation := input.Id == ""
	c.SetLogger(c.Logger().With(
		"aircraft", input.Aircraft,
		"pilot", input.Pilot,
		"type", input.ReservationType,
		"start", input.StartTime.Format("2006-01-02T15:04:05.000Z"),
		"end", input.EndTime.Format("2006-01-02T15:04:05.000Z")))
	if input.Instructor != nil {
		c.SetLogger(c.Logger().With("instructor", *input.Instructor))
	}

	if newReservation {
		c.Logger().Info("creating reservation")
	} else {
		c.SetLogger(c.Logger().With("reservation", input.Id))
		c.Logger().Info("updating reservation")
	}

	// Check invalid times
	if input.StartTime == input.EndTime {
		return nil, ReservationTimesEqualError
	}

	if input.EndTime.Compare(input.StartTime) < 0 {
		return nil, ReservationTimesSwappedError
	}

	if newReservation {
		input.Id = ulid.Make().String()
		c.SetLogger(c.Logger().With("reservation", input.Id))

		if input.StartTime.Compare(time.Now()) < 0 {
			return nil, ReservationCreateTimePastError
		}
	} else if time.Now().Sub(input.EndTime).Seconds() > 0 {
		return nil, ReservationPastUpdateError
	}

	databaseItem := databaseItem{
		PK:          fmt.Sprintf("%s#%s", CLUB_PARTITION_KEY, CLUB_ID),
		SK:          fmt.Sprintf("%s#%s", RESERVATION_PARTITION_KEY, input.Id),
		ItemType:    "reservation",
		Reservation: input,
	}

	out, err := c.DatabaseClient.Put(database.PutInput{Item: databaseItem})
	if err != nil {
		return nil, err
	}

	input.CreatedAt = out.CreatedAt
	input.UpdatedAt = out.UpdatedAt

	if newReservation {
		c.Logger().Info("reservation created")
	} else {
		c.Logger().Info("reservation updated")
	}

	return &input, nil
}

type ListInput struct {
	NextToken *string
	Limit     *int32
}

type ListOutput struct {
	NextToken *string       `json:"nextToken"`
	Results   []Reservation `json:"results"`
}

type ExclusiveStartKey struct {
	PK string `json:"PK"`
	SK string `json:"SK"`
}

// Returns stored data for all reservations.
func (c *Client) List(input ListInput) (*ListOutput, error) {
	c.Logger().Info("listing reservations")

	var keyConditionExpression *string
	var expressionAttributeValues = make(map[string]types.AttributeValue)

	pk := fmt.Sprintf("%s#%s", CLUB_PARTITION_KEY, CLUB_ID)
	var sk string

	keyConditionExpression = aws.String("PK = :pk AND begins_with(SK, :sk)")
	sk = RESERVATION_PARTITION_KEY
	expressionAttributeValues[":sk"] = &types.AttributeValueMemberS{
		Value: sk,
	}
	expressionAttributeValues[":pk"] = &types.AttributeValueMemberS{
		Value: pk,
	}

	var exclusiveStartKey map[string]types.AttributeValue
	if input.NextToken != nil {
		nextTokenData, err := base64.StdEncoding.DecodeString(*input.NextToken)
		if err != nil {
			return nil, err
		}

		var nextToken map[string]string
		err = json.Unmarshal([]byte(nextTokenData), &nextToken)
		if err != nil {
			return nil, err
		}

		nextTokenMap, err := attributevalue.MarshalMap(nextToken)
		if err != nil {
			return nil, err
		}

		exclusiveStartKey = nextTokenMap
	}

	queryInput := database.QueryInput{
		Limit:                     input.Limit,
		KeyConditionExpression:    keyConditionExpression,
		ExpressionAttributeValues: expressionAttributeValues,
		ExclusiveStartKey:         exclusiveStartKey,
	}

	output, err := c.DatabaseClient.Query(&queryInput)
	if err != nil {
		return nil, err
	}

	var reservations = make([]Reservation, 0)
	for _, item := range output.Items {
		var reservation Reservation
		err := attributevalue.UnmarshalMap(item, &reservation)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}

	var nextToken *string
	if len(output.LastEvaluatedKey) > 0 {
		token := new(ExclusiveStartKey)
		err = attributevalue.UnmarshalMap(output.LastEvaluatedKey, token)
		if err != nil {
			return nil, err
		}

		jsonOut, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}

		nextToken = aws.String(string(base64.StdEncoding.EncodeToString(jsonOut)))
	}

	if err != nil {
		return nil, err
	} else {
		c.Logger().Info("reservations listed", "count", len(reservations), "isNextToken", nextToken != nil)
		return &ListOutput{
			NextToken: nextToken,
			Results:   reservations,
		}, nil
	}
}

// Returns stored data for a reservation.
func (c *Client) Get(reservationId string) (*Reservation, error) {
	c.SetLogger(c.Logger().With("reservation", reservationId))
	c.Logger().Info("retrieving reservation")

	output, err := c.DatabaseClient.Get(database.GetInput{
		PK: fmt.Sprintf("%s#%s", CLUB_PARTITION_KEY, CLUB_ID),
		SK: fmt.Sprintf("%s#%s", RESERVATION_PARTITION_KEY, reservationId),
	})
	if err != nil {
		return nil, err
	}

	reservation := new(Reservation)
	err = attributevalue.UnmarshalMap(output.Item, reservation)
	if err != nil {
		return nil, err
	}

	c.Logger().Info("reservation retrieved")
	return reservation, nil
}

func (c *Client) Delete(reservationId string) error {
	c.SetLogger(c.Logger().With("reservation", reservationId))
	c.Logger().Info("deleting reservation")

	_, err := c.DatabaseClient.Delete(&database.DeleteInput{
		PK: fmt.Sprintf("%s#%s", CLUB_PARTITION_KEY, CLUB_ID),
		SK: fmt.Sprintf("%s#%s", RESERVATION_PARTITION_KEY, reservationId),
	})

	if err != nil {
		c.Logger().Info("reservation deleted")
	}

	return err
}
