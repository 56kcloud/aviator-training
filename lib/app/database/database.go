/*
Package database abstracts the various AWS DynamoDB APIs by providing interfaces and methods.
*/
package database

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDbAPI interface {
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type Config struct {
	DynamoDbClient DynamoDbAPI
	TableName      string
}

type Client struct {
	Config
}

// Returns a new database API client from the provided config.
func NewFromConfig(c Config) *Client {
	return &Client{Config: c}
}

type Timestamp struct {
	CreatedAt time.Time
	UpdateAt  time.Time
}

type TransactWriteOutput struct {
	Output     *dynamodb.TransactWriteItemsOutput
	Timestamps []Timestamp
}

func (c Client) TransactWriteItems(params dynamodb.TransactWriteItemsInput) (*TransactWriteOutput, error) {
	timestamps := make([]Timestamp, 0)
	for i := 0; i < len(params.TransactItems); i++ {
		if params.TransactItems[i].Put != nil {
			currentTime := time.Now()
			params.TransactItems[i].Put.Item["CreatedAt"] = &types.AttributeValueMemberS{
				Value: currentTime.Format("2006-01-02T15:04:05.000Z"),
			}
			params.TransactItems[i].Put.Item["UpdatedAt"] = &types.AttributeValueMemberS{
				Value: currentTime.Format("2006-01-02T15:04:05.000Z"),
			}
			timestamps = append(timestamps, Timestamp{
				CreatedAt: currentTime,
				UpdateAt:  currentTime,
			})
		}
	}
	output, err := c.DynamoDbClient.TransactWriteItems(context.TODO(), &params)
	return &TransactWriteOutput{
		Output:     output,
		Timestamps: timestamps,
	}, err
}

type BatchWriteInput struct {
	Items []map[string]types.AttributeValue
}

func (c Client) BatchWriteItem(input BatchWriteInput) error {
	requests := make([]types.WriteRequest, 0)
	for i := 0; i < len(input.Items); i++ {
		item := input.Items[i]
		currentTime := time.Now()
		item["CreatedAt"] = &types.AttributeValueMemberS{
			Value: currentTime.Format("2006-01-02T15:04:05.000Z"),
		}
		item["UpdatedAt"] = &types.AttributeValueMemberS{
			Value: currentTime.Format("2006-01-02T15:04:05.000Z"),
		}

		requests = append(requests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: item,
			},
		})
	}

	requestMap := make(map[string][]types.WriteRequest)
	requestMap[c.TableName] = requests

	_, err := c.DynamoDbClient.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: requestMap,
	})
	return err
}

type BatchDeleteInput struct {
	Items []map[string]types.AttributeValue
}

func (c Client) BatchDeleteItem(input BatchDeleteInput) error {
	requests := make([]types.WriteRequest, 0)
	for i := 0; i < len(input.Items); i++ {
		item := input.Items[i]

		requests = append(requests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"PK": item["PK"],
					"SK": item["SK"],
				},
			},
		})
	}

	requestMap := make(map[string][]types.WriteRequest)
	requestMap[c.TableName] = requests

	_, err := c.DynamoDbClient.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: requestMap,
	})
	return err
}

type PutInput struct {
	Item                     any
	ConditionExpression      *string
	ExpressionAttributeNames map[string]string
}

type PutOutput struct {
	Ouput     *dynamodb.PutItemOutput
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c Client) Put(input PutInput) (*PutOutput, error) {
	item, err := attributevalue.MarshalMap(input.Item)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	item["CreatedAt"] = &types.AttributeValueMemberS{
		Value: currentTime.Format("2006-01-02T15:04:05.000Z"),
	}
	item["UpdatedAt"] = &types.AttributeValueMemberS{
		Value: currentTime.Format("2006-01-02T15:04:05.000Z"),
	}

	output, err := c.DynamoDbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:                &c.TableName,
		Item:                     item,
		ConditionExpression:      input.ConditionExpression,
		ExpressionAttributeNames: input.ExpressionAttributeNames,
	})
	return &PutOutput{
		Ouput:     output,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}, err
}

type GetInput struct {
	PK string
	SK string
}

func (c Client) Get(input GetInput) (*dynamodb.GetItemOutput, error) {
	keyItem, err := attributevalue.MarshalMap(input)
	if err != nil {
		return nil, err
	}
	output, err := c.DynamoDbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &c.TableName,
		Key:       keyItem,
	})
	return output, err
}

type UpdateInput struct {
	PK                  string
	SK                  string
	Item                any
	ConditionExpression *string
}

func UpdateExpression(input any, setNil bool) (expression.Expression, time.Time, error) {
	fields := reflect.TypeOf(input)
	v := reflect.ValueOf(input)

	upd := expression.UpdateBuilder{}
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).IsNil() {
			valueType := v.Field(i).Elem().Type()
			valueBuilder := expression.ValueBuilder{}

			switch valueType.String() {
			case "string":
				valueBuilder = expression.Value(v.Field(i).Elem().String())
			case "int":
				valueBuilder = expression.Value(v.Field(i).Elem().Int())
			case "float32":
				valueBuilder = expression.Value(v.Field(i).Elem().Float())
			case "bool":
				valueBuilder = expression.Value(v.Field(i).Elem().Bool())
			default:
				valueBuilder = expression.Value(v.Field(i).Elem().Interface())
			}

			if fields.Field(i).Name == "GSIData" {
				for k := range v.Field(i).Elem().Interface().(map[string]interface{}) {
					mapV := v.Field(i).Elem().Interface().(map[string]interface{})[k]
					valueBuilder = expression.Value(mapV)
					upd = upd.Set(expression.Name(fmt.Sprintf("%s.%s", "GSIData", k)), valueBuilder)
				}
			} else {
				upd = upd.Set(expression.Name(fields.Field(i).Name), valueBuilder)
			}
		} else {
			if setNil {
				valueBuilder := expression.ValueBuilder{}
				upd = upd.Set(expression.Name(fields.Field(i).Name), valueBuilder)
			}
		}
	}
	currentTime := time.Now()
	upd = upd.Set(expression.Name("UpdatedAt"), expression.Value(currentTime.Format("2006-01-02T15:04:05.000Z")))

	expr, err := expression.NewBuilder().WithUpdate(upd).Build()
	return expr, currentTime, err
}

type UpdateOutput struct {
	Ouput     *dynamodb.UpdateItemOutput
	UpdatedAt time.Time
}

func (c Client) Update(input UpdateInput) (*UpdateOutput, error) {
	expr, updatedAt, err := UpdateExpression(input.Item, false)
	if err != nil {
		return nil, err
	}

	output, err := c.DynamoDbClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: &c.TableName,
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: input.PK,
			},
			"SK": &types.AttributeValueMemberS{
				Value: input.SK,
			},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       input.ConditionExpression,
	})
	return &UpdateOutput{
		Ouput:     output,
		UpdatedAt: updatedAt,
	}, err
}

type QueryInput struct {
	Index                     *string
	KeyConditionExpression    *string
	ExpressionAttributeValues map[string]types.AttributeValue
	ExpressionAttributeNames  map[string]string
	Limit                     *int32
	ExclusiveStartKey         map[string]types.AttributeValue
	FilterExpression          *string
	ScanIndexForward          *bool
}

func (c Client) Query(queryInput *QueryInput) (*dynamodb.QueryOutput, error) {
	queryOutput, err := c.DynamoDbClient.Query(
		context.TODO(),
		&dynamodb.QueryInput{
			TableName:                 &c.TableName,
			IndexName:                 queryInput.Index,
			KeyConditionExpression:    queryInput.KeyConditionExpression,
			ExpressionAttributeValues: queryInput.ExpressionAttributeValues,
			ExpressionAttributeNames:  queryInput.ExpressionAttributeNames,
			Limit:                     queryInput.Limit,
			ExclusiveStartKey:         queryInput.ExclusiveStartKey,
			FilterExpression:          queryInput.FilterExpression,
			ScanIndexForward:          queryInput.ScanIndexForward,
		})

	return queryOutput, err
}

type DeleteInput struct {
	PK string
	SK string
}

func (c Client) Delete(deleteInput *DeleteInput) (*dynamodb.DeleteItemOutput, error) {
	deleteOutput, err := c.DynamoDbClient.DeleteItem(
		context.TODO(),
		&dynamodb.DeleteItemInput{
			TableName: &c.TableName,
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: deleteInput.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: deleteInput.SK,
				},
			},
		})

	return deleteOutput, err
}
