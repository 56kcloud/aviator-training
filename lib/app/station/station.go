package station

import (
	"app/constants"
	"app/database"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/oklog/ulid/v2"
)

type Station struct {
	Id        string
	Longitude string
	Latitude  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type databaseItem struct {
	PK     string
	SK     string
	GSI1PK string
	GSI1SK string
	Station
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

func (c Client) Create(input *Station) (*Station, error) {
	input.Id = ulid.Make().String()
	databaseItem := databaseItem{
		PK:       fmt.Sprintf("%s#%s", constants.STATION_PARTITION_KEY, input.Id),
		SK:       fmt.Sprintf("%s#%s", constants.STATION_PARTITION_KEY, input.Id),
		GSI1PK:   constants.STATION_PARTITION_KEY,
		GSI1SK:   fmt.Sprintf("%s#%s", constants.STATION_PARTITION_KEY, input.Id),
		Station:  *input,
		ItemType: "station",
	}
	_, err := c.DatabaseClient.Put(database.PutInput{
		Item: databaseItem,
	})

	if err != nil {
		return nil, err
	}

	return input, nil
}

func (c Client) Get(stationId string) (*Station, error) {
	output, err := c.DatabaseClient.Get(database.GetInput{
		PK: fmt.Sprintf("%s#%s", constants.STATION_PARTITION_KEY, stationId),
		SK: fmt.Sprintf("%s#%s", constants.STATION_PARTITION_KEY, stationId),
	})
	if err != nil {
		return nil, err
	}

	item := new(Station)
	err = attributevalue.UnmarshalMap(output.Item, item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (c Client) List() (*[]Station, error) {
	queryInput := database.QueryInput{
		KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{
				Value: constants.STATION_PARTITION_KEY,
			},
			":sk": &types.AttributeValueMemberS{
				Value: constants.STATION_PARTITION_KEY,
			},
		},
	}

	output, err := c.DatabaseClient.Query(&queryInput)
	if err != nil {
		return nil, err
	}

	var items = make([]Station, 0)
	for _, item := range output.Items {
		station := new(Station)
		err := attributevalue.UnmarshalMap(item, station)
		if err != nil {
			return nil, err
		}

		items = append(items, *station)
	}

	return &items, nil
}
