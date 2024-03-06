# Aviator Training

Welcome to Aviator Training, a small serverless AWS backend that allows pilots to reserve aircraft and developers to learn serverless.

## Setup

### Create a Cloud9 environment
In the AWS console search bar, type Cloud9 and select the corresponding service in the list.

- Click the orange **Create environment** button.
- For name: **workshop**
- Click the **Additional instance types** box and in the dropdown menu select **t3.medium** as the instance type.
- Click **Create**
- Back in the list of environment open the **workshop** environment you just created.

### Clone this repository
In your Cloud9 environment, open a terminal and run:
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
cd aviator-training
sh deploy.sh -s organization/aviator/dev
```

You will be asked if you want to create the "dev" stack:
```
If you would like to create this stack now, please press <ENTER>, otherwise press ^C: 
```
Press enter.
