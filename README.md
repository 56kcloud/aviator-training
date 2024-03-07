# Aviator Training

Welcome to Aviator Training, a small serverless AWS backend written in Go that allows pilots to reserve aircraft and developers to learn about serverless cloud architectures. Using Infrastructure as Code you will provision a database, a Lambda function with the business logic code and a REST API.

## Prerequesite
You are signed into an AWS Lab environment and have access to the following screen:
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/access.png)
Use the button on the left to access your own AWS console environment.

## Setup

### Create a Cloud9 environment
In the AWS console search bar, type Cloud9 and select the corresponding service in the list.

- Click the orange **Create environment** button.
- For name: **workshop**
- Click the **Additional instance types** box and in the dropdown menu select **t3.medium** as the instance type.
- Click **Create**
- Back in the list of environment open the **workshop** environment you just created.

  
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/cloud9-basic-env.png)
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/create-cloud9-env.png)

### Replace default AWS credentials
Cloud9 comes with its own managed credentials by default. However these do not have the necessary permissions for this workshop. So let's turn them off and replace them.

1. In the top right corner, click the Cog (settings) button -> AWS Settings -> unselect *AWS managed temporary credentials*.
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/cloud9-aws-settings.png)
2. In the terminal, Copy & Paste the credentials received via this lab environment.

![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/access-cli-credentials.png)
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/cli-credentials.png)
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/paste-cli-credentials.png)

### Clone this repository
In your Cloud9 environment, run the following in the bottom terminal:
```
git clone https://github.com/56kcloud/aviator-training.git
cd aviator-training
```

### Install Pulumi and latest Go version
Copy and paste the following command and press enter:
```
source setup.sh
```

### Configure Pulumi
Pulumi, like all Infrastructure as Code frameworks, needs to be able to track the state of deployed infrastructure. When working with AWS, this state information can be stored in an S3 bucket. Your Cloud9 environment is loaded with the AWS CLI (Command Line Interface). Let's use to create an S3 bucket that we will use to store our state files:
> **Warning**
> S3 bucket names must be unique. Replace xxxxx by a random letters of your choice. For example sdlk1.
> 
```
aws s3api create-bucket --bucket pulumi-state-xxxxx --region eu-west-1 --create-bucket-configuration LocationConstraint=eu-west-1
```

Tell Pulumi to use this bucket, again replacing xxxxx.
```
pulumi login s3://pulumi-state-xxxxx
```

## Deploy the app
```
sh deploy.sh -s organization/aviator/dev
```

You will be asked if you want to create the "dev" stack:
```
The stack 'organization/aviator/dev' does not exist.
If you would like to create this stack now, please press <ENTER>, otherwise press ^C: 
```
Press enter.

The first time you run this command, it will take **around 2 minutes** to install all the application dependencies. When all said and done, you should see the following output:
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/initial-pulumi-output.png)

Run the previous command again, you will see that Pulumi is stateful, it will not provision any new resources.

## Calling the API
The output of the Pulumi command provides a **Reservation API url** that looks like *https://[some-id].execute-api.eu-west-1.amazonaws.com/v1/reservations*. Click that URL -> Open. You should see the following JSON output in a tab: 
```
{"nextToken":null,"results":[]}
```
Great, the API returns a successfull response. It is empty because we have not created any reservations yet. Let's change that now. To create a reservation by making an HTTP POST request against your API, run the following command after replacing the [api-id] with yours:
```
curl --location 'https://[api-id].execute-api.eu-west-1.amazonaws.com/v1/reservations' \
--header 'Content-Type: application/json' \
--data '{
    "aircraft": "HB-KFQ",
    "reservationType": "Sightseeing",
    "startTime": "2024-04-07T16:00:00Z",
    "endTime": "2024-04-07T17:00:00Z",
    "pilot": "Jane Doe",
    "remarks": ""
}'
```
If you still have the tab from the previous step open, refresh it and you will see it now returns the reservation you just created.

## Under the hood
Now that we have deployed this app, let's take a look at what was deployed. This application uses three main AWS managed services:
- API Gateway to create a REST API
- A single Lambda function with our business logic
- A DynamoDB table to store the reservations

![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/architecture.png)

Before diving into the code, let's visit these services in the AWS console.

### API Gateway
In the [API Gateway console](https://eu-west-1.console.aws.amazon.com/apigateway/main/apis?region=eu-west-1), you should see an `aviator-rest-api` in the list of APIs. If you open it you will find a number of of API routes. For example if you click the `/reservations - GET` resource, you will find that this API route is mapped to a Lambda function.
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/api-gateway.png)

### Lambda
In the [Lambda console](https://eu-west-1.console.aws.amazon.com/lambda/home?region=eu-west-1#/functions), there is a function called "app". Open it and you will find the configuration details of the function: 
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/lambda.png)

### DynamoDB
In the [DynamoDB console](https://eu-west-1.console.aws.amazon.com/dynamodbv2/home?region=eu-west-1#tables), there is a table called "aviator-table". Open it and you will find the configuration and monitoring details of the table: 
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/dynamodb-table.png)
If you navigate on the left menu to "Explore items", you can see the items stored in your table:
![Alt](https://github.com/56kcloud/aviator-training/blob/doc/doc/img/dynamodb-explore.png)

### How was this deployed
If we now go back to the code, let's look what's inside:
```
├── api.json # OpenAPI JSON spec
├── deploy.sh # Command used to compile the Go code and provision the resources
├── cmd 
│   ├── functions # AWS Lambda functions
│   │   ├── app # Main "app" function
│   └── infrastructure # Go command to start Pulumi
├── lib
│   └── aviator # Business logic Go package imported by the Lambda function
│   ├── infrastructure # Pulumi resource provisioning code
│   │   ├── api # Provision API Gateway and Lambda integration
│   │   ├── database # Provision the DynamoDB table
```

To get an idea of what Infrastructure as Code looks like, open `lib/infrastructure/database/database.go`. In this file you'll see we are using the Pulumi AWS SDK to create a new DynamoDB table by simply providing configuration properties (table name, Hash key, Range key, etc...). The list of configuration properties are of course provided by the [Pulumi documentation](https://www.pulumi.com/registry/packages/aws/api-docs/dynamodb/table/), itself backed by the [official AWS documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html).

Ok, now let's look at a Lambda function. Open `cmd/functions/app/main.go`. At the  bottom of the file you'll find:
```
func main() {
  lambda.Start(HandleRequest)
}
```

This is the entry point, when this function executes on AWS, the Lambda service will execute our Go executable and the Lambda will pass the trigger event to the `HandleRequest` function located in the middle of the file. Notice the signature of the `HandleRequest` function:
```
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
```
It takes as second parameter an event of type `APIGatewayProxyRequest`. This event contains everything we need to know to handle the incoming API request: the API route, HTTP method (GET, POST, etc...), JSON payload, etc... Once we have this event the rest is our own business logic.

In this same file at the top you'll see we import notable the DynamoDB AWS SDK (`github.com/aws/aws-sdk-go-v2/service/dynamodb`) and our own reservation package (`aviator/reservation`). The code for the latter you can find in `lib/aviator/reservation/reservation.go`. But staying within the Lambda code for now, within the `HandleRequest` function the code initializes the a Database and Reservation clients. Finally if the incoming request concerns reservations, it calls the `reservationCrud` function (code can be found in `cmd/functions/app/reservation.go`), which is just a "switch case" function that calls the correct reservation library method based on whether we want to create, retrieve, update or delete (CRUD) a reservation:
```
if strings.HasPrefix(path, "/reservations") {
    reservationClient.SetLogger(logger)
    return reservationCrud(ctx, request, path, stage, reservationClient, *errorClient)
}
```
The response of `reservationCrud` is returned to the Lambda and is of type `APIGatewayProxyResponse` which API Gateway can return to the caller. 

Ok, but how was API Gateway configured with the reservation API? In the route of this project, you'll find an `api.json`. These contains an [OpenAPI](https://swagger.io/specification/) spec. OpenAPI is an open-source and widely used API specification format. API Gateway is able to consume this JSON file and automatically configure itself based on its contents. Backed to the Pulumi code this, as well as the Lambda function provisioning, is done in the the`lib/infrastructure/api/api.go` file.

### Roles and permissions
AWS works on a "by default no access is given" basis. This is valid also for the resources and code you deploy. By default your code has no access to other AWS services, access must be given. To demonstrate this let's break our app 🙂. Open `lib/infrastructure/api/api.go` and look for the function:
```go
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
									"dynamodb:GetItem",
									"dynamodb:PutItem",
									"dynamodb:DeleteItem",
									"dynamodb:Query",
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
```
This is the defination of an IAM (Identity and Access Management) role, a building block of AWS. In this case, Pulumi attaches this role to the Lambda function, without it the code that runs within the Lambda has no access to other AWS services. In our case we see that the current role statement provides access to DynamoDB (the `dynamodb:GetItem`, etc... statements). Let's change this role and see what happens. Remove all actions except the `dynamodb:DeleteItem` and save the file (if you get a warning about code formatting, ignore it):
```go
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
									"dynamodb:DeleteItem",
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
```

Run the deploy command:
```
sh deploy.sh -s organization/aviator/dev
```

The Pulumi output should show that it updated the `lambda-execution-role` resource.

Let's call our API and see what happens (remember to replace the API ID):
```
curl https://xxxxxxx.execute-api.eu-west-1.amazonaws.com/v1/reservations
```

You should see that no reservations are returned. Instead the response is:
```
{"message":"Unauthorized"}
```
The Lambda function has no READ access to the database and can't fetch the data.

You can put the original IAM role definition back, save, deploy and test again.

### Destroying the app
Resources created by Pulumi can also be deleted. Let's remove the entire app:
```
sh destroy.sh -s organization/aviator/dev
```
You will be asked to confirm, select `yes`.

If you try to call the API again, it won't work! Everything is gone.

## Conclusion
This small workshop demonstrated how easy it is to get going on the AWS cloud with serverless architectures. The services used today were all managed by AWS. They can be provisioned and destroyed within seconds. There was no need to manage them, we only needed to configure them. This gives developers time to focus on the code and business logic instead. Furthermore everything we used today was on-demand, meaning we only pay for the requests we served. There were no servers running 24 / 7: if our app has no traffic nothing is running.
