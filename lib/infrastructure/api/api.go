package api

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const LAMBDA_MEMORY_SIZE = 768

type Input struct {
	OpenAPISpecPath   string
	DynamodbTableName string
	DynamodbTableArn  pulumi.StringOutput
}

type Output struct {
	Api                     *apigateway.RestApi
	ProdStageName           pulumi.StringOutput
	TestStageName           pulumi.StringOutput
	ApplicationFunctionName string
}

func Provision(ctx *pulumi.Context, input Input) (*Output, error) {
	_, err := lambdaInvocationRole(ctx)
	if err != nil {
		return nil, err
	}
	lambdaExecutionRole, err := lambdaExecutionRole(ctx, input)
	if err != nil {
		return nil, err
	}

	layer, err := lambda.LookupLayerVersion(ctx, &lambda.LookupLayerVersionArgs{
		LayerName: "arn:aws:lambda:eu-west-1:901920570463:layer:aws-otel-collector-arm64-ver-0-68-0",
		Version:   pulumi.IntRef(1),
	})
	if err != nil {
		return nil, err
	}

	rootDir := "../.."

	appFunctionName := "app"
	appFunction(ctx, appFunctionName, rootDir, input, lambdaExecutionRole, layer.Arn)

	apiSpec, err := os.ReadFile(input.OpenAPISpecPath)
	if err != nil {
		return nil, err
	}
	apiName := "aviator-rest-api"
	api, err := apigateway.NewRestApi(ctx, apiName, &apigateway.RestApiArgs{
		Name: pulumi.String(apiName),
		Body: pulumi.String(apiSpec),
		EndpointConfiguration: apigateway.RestApiEndpointConfigurationArgs{
			Types: pulumi.String("REGIONAL"),
		},
	})
	if err != nil {
		return nil, err
	}

	deployment, err := apigateway.NewDeployment(ctx, "aviator-rest-api-deployment", &apigateway.DeploymentArgs{
		RestApi: api.ID(),
		// Any update to the OpenAPI JSON spec should trigger a new deployment
		Triggers: pulumi.StringMap{
			"redeployment": pulumi.String(sha1Hash(string(apiSpec))),
		},
	})
	if err != nil {
		return nil, err
	}

	prodStage, err := apigateway.NewStage(ctx, "aviator-rest-api-v1-stage", &apigateway.StageArgs{
		StageName:  pulumi.String("v1"),
		RestApi:    api,
		Deployment: deployment,
		Variables: pulumi.StringMap{
			"name": pulumi.String("v1"),
		},
	})
	if err != nil {
		return nil, err
	}

	testStage, err := apigateway.NewStage(ctx, "aviator-rest-api-test-stage", &apigateway.StageArgs{
		Description: pulumi.String("Stage test is used for API contract testing only"),
		StageName:   pulumi.String("test"),
		RestApi:     api,
		Deployment:  deployment,
		Variables: pulumi.StringMap{
			"name": pulumi.String("test"),
		},
	})
	if err != nil {
		return nil, err
	}

	return &Output{
		Api:                     api,
		ProdStageName:           prodStage.StageName,
		TestStageName:           testStage.StageName,
		ApplicationFunctionName: appFunctionName,
	}, err
}

func sha1Hash(input string) string {
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

func lambdaExecutionRole(ctx *pulumi.Context, input Input) (*iam.Role, error) {
	roleName := "lambda-execution-role"
	role, err := iam.NewRole(ctx, roleName, &iam.RoleArgs{
		Name: pulumi.String(roleName),
		ManagedPolicyArns: pulumi.StringArray{
			iam.ManagedPolicyCloudWatchLogsFullAccess,
			iam.ManagedPolicyAWSXRayDaemonWriteAccess,
		},
		AssumeRolePolicy: pulumi.String(`{
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
				"Action": "sts:AssumeRole",	
                "Principal": { "Service": "lambda.amazonaws.com" }
            }]
       }`),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name: pulumi.String("dynamodb-access-policy"),
				Policy: input.DynamodbTableArn.ApplyT(func(arn string) string {
					document, _ := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
						Statements: []iam.GetPolicyDocumentStatement{
							{
								Effect: pulumi.StringRef("Allow"),
								Actions: []string{
									"dynamodb:*",
								},
								Resources: []string{
									arn,
								},
							},
						},
					})
					return document.Json
				}).(pulumi.StringOutput),
			},
		},
	})
	return role, err
}

// IAM role that allows API Gateway to call our Lambda
func lambdaInvocationRole(ctx *pulumi.Context) (*iam.Role, error) {
	current, err := aws.GetCallerIdentity(ctx, nil)
	if err != nil {
		return nil, err
	}

	region, err := aws.GetRegion(ctx, nil)
	if err != nil {
		return nil, err
	}

	roleName := "api-gateway-invoke-lambda-role"
	role, err := iam.NewRole(ctx, roleName, &iam.RoleArgs{
		Name: pulumi.String(roleName),
		AssumeRolePolicy: pulumi.String(`{
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
				"Action": "sts:AssumeRole",	
                "Principal": { "Service": "apigateway.amazonaws.com" }
            }]
        }`),
	})
	if err != nil {
		return nil, err
	}

	arnPrefix := fmt.Sprintf("arn:aws:lambda:%s:%s:function:", region.Name, current.AccountId)

	invokeLambdaPolicyDocument, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: pulumi.StringRef("Allow"),
				Actions: []string{
					"lambda:InvokeFunction",
				},
				Resources: []string{
					arnPrefix + "app",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	policyName := "invoke-lambda-policy"
	invokeLambdaPolicy, err := iam.NewPolicy(ctx, policyName, &iam.PolicyArgs{
		Name:   pulumi.String(policyName),
		Policy: pulumi.String(invokeLambdaPolicyDocument.Json),
	})
	if err != nil {
		return nil, err
	}

	_, err = iam.NewPolicyAttachment(ctx, "invoke-lambda-policy-attachment", &iam.PolicyAttachmentArgs{
		PolicyArn: invokeLambdaPolicy.Arn,
		Roles: pulumi.Array{
			role.Name,
		},
	})

	return role, err
}

func appFunction(ctx *pulumi.Context, functionName string, rootDir string, input Input, role *iam.Role, otelLayerArn string) error {
	environment := pulumi.All(input.DynamodbTableName).ApplyT(
		func(args []interface{}) pulumi.StringMap {
			return pulumi.StringMap{
				"DYNAMODB_TABLE_NAME": pulumi.String(args[0].(string)),
			}
		},
	).(pulumi.StringMapOutput)

	_, err := lambda.NewFunction(ctx, functionName, &lambda.FunctionArgs{
		Name:    pulumi.String(functionName),
		Runtime: pulumi.String("provided.al2"),
		Architectures: pulumi.StringArray{
			pulumi.String("arm64"),
		},
		MemorySize:    pulumi.Int(LAMBDA_MEMORY_SIZE),
		Timeout:       pulumi.Int(30),
		Code:          pulumi.NewFileArchive(rootDir + "/cmd/functions/app/."),
		Handler:       pulumi.String("bootstrap"),
		Role:          role.Arn,
		TracingConfig: lambda.FunctionTracingConfigArgs{Mode: pulumi.String("Active")},
		Publish:       pulumi.Bool(true),
		Layers: pulumi.StringArray{
			pulumi.String(otelLayerArn),
		},
		Environment: lambda.FunctionEnvironmentArgs{
			Variables: environment,
		},
	})

	return err
}
