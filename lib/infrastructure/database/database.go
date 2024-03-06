package database

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TableOutputs struct {
	TableName string
	TableArn  pulumi.StringOutput
}

func Provision(ctx *pulumi.Context) (*TableOutputs, error) {
	name := "aviator-table"
	table, err := dynamodb.NewTable(ctx, name, &dynamodb.TableArgs{
		Name:     pulumi.String(name),
		HashKey:  pulumi.String("PK"),
		RangeKey: pulumi.String("SK"),
		Attributes: dynamodb.TableAttributeArray{
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("PK"),
				Type: pulumi.String("S"),
			},
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("SK"),
				Type: pulumi.String("S"),
			},
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("GSI1PK"),
				Type: pulumi.String("S"),
			},
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("GSI1SK"),
				Type: pulumi.String("S"),
			},
		},
		GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
			&dynamodb.TableGlobalSecondaryIndexArgs{
				Name:           pulumi.String("GSI1"),
				HashKey:        pulumi.String("GSI1PK"),
				RangeKey:       pulumi.String("GSI1SK"),
				ProjectionType: pulumi.String("INCLUDE"),
				NonKeyAttributes: pulumi.StringArray{
					pulumi.String("CreatedAt"),
					pulumi.String("UpdatedAt"),
					pulumi.String("GSIData"),
					pulumi.String("Id"),
					pulumi.String("ItemType"),
				},
			},
		},
		BillingMode: pulumi.String("PAY_PER_REQUEST"),
	})

	return &TableOutputs{
		TableName: name,
		TableArn:  table.Arn,
	}, err
}
