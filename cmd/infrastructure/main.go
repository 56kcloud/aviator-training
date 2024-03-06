package main

import (
	"infrastructure/database"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		database.ProvisionTable(ctx)
		return nil
	})
} 
