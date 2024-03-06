package measurement

import (
	"app/constants"
	"app/database"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

type Measurement struct {
	StationId   string
	Temperature int
	Barometer   int
	MeasuredAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type databaseItem struct {
	PK string
	SK string
	Measurement
	ItemType string
}

type Config struct {
	DatabaseClient database.Client
}

type Client struct {
	Config
}

func NewFromConfig(c Config) *Client {
	return &Client{Config: c}
}

func (c Client) Create(input *Measurement) (*Measurement, error) {
	databaseItem := databaseItem{
		PK:          fmt.Sprintf("%s#%s", constants.STATION_PARTITION_KEY, input.StationId),
		SK:          fmt.Sprintf("%s#%s", constants.MEASUREMENT_PARTITION_KEY, input.MeasuredAt),
		Measurement: *input,
		ItemType:    "measurement",
	}
	_, err := c.DatabaseClient.Put(database.PutInput{
		Item: databaseItem,
	})

	if err != nil {
		return nil, err
	}

	return input, nil
}

func (c Client) List(stationId string) (*[]Measurement, error) {
	queryInput := database.QueryInput{
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{
				Value: fmt.Sprintf("%s#%s", constants.STATION_PARTITION_KEY, stationId),
			},
			":sk": &types.AttributeValueMemberS{
				Value: constants.MEASUREMENT_PARTITION_KEY,
			},
		},
	}

	output, err := c.DatabaseClient.Query(&queryInput)
	if err != nil {
		return nil, err
	}

	var items = make([]Measurement, 0)
	for _, item := range output.Items {
		measurement := new(Measurement)
		err := attributevalue.UnmarshalMap(item, measurement)
		if err != nil {
			return nil, err
		}

		items = append(items, *measurement)
	}

	return &items, nil
}
