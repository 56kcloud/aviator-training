package main

import (
	"infrastructure/api"
	"infrastructure/database"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		databaseOut, err := database.Provision(ctx)
		if err != nil {
			return err
		}

		apiOut, err := api.Provision(ctx, api.Input{
			OpenAPISpecPath:   "../../api.json",
			DynamodbTableName: databaseOut.TableName,
			DynamodbTableArn:  databaseOut.TableArn,
		})
		if err != nil {
			return err
		}

		regionOut, err := aws.GetRegion(ctx, nil, nil)
		if err != nil {
			return err
		}

		ctx.Export("Reservation API url", pulumi.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s/reservations", apiOut.Api.ID(), regionOut.Name, apiOut.ProdStageName))
		return nil
	})
}
