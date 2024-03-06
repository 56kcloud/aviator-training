# Aviator Training

Welcome to Aviator Training, a small serverless AWS backend that allows pilots to reserve aircraft and developers to learn serverless.

## Setup

### Clone this repository
In your Cloud9 environment, open a terminal and run:
```
git clone https://github.com/56kcloud/aviator-training.git
```

### Install Pulumi
Open a terminal in Cloud9, copy and paste the following command and press enter:
```
curl -fsSL https://get.pulumi.com | sh
```

Close the terminal and open a new one and verify a successfull installation:
```
pulumi version
```
The command should return `v3`.

### Configure Pulumi
Pulumi like all Infrastructure as Code frameworks, needs to be able to track the state of deployed infrastructure. When working with AWS, this state information can be stored in an S3 bucket. Your Cloud9 environment is loaded with the AWS CLI (Command Line Interface). Let's use to create an S3 bucket that we will use to store our state files:
> **Warning**
> S3 bucket names must be unique. Replace xxxxx by a random letters of your choice. For example sdlkm1
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

You will be asked to set a password for secrets:
```
Enter your passphrase to unlock config/secrets
    (set PULUMI_CONFIG_PASSPHRASE or PULUMI_CONFIG_PASSPHRASE_FILE to remember):  
```
We don't need this so just press enter.

```
export PULUMI_CONFIG_PASSPHRASE=
```
