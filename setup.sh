#!/bin/bash

curl -fsSL https://get.pulumi.com | sh
export PATH=$PATH:$HOME/.pulumi/bin
export PULUMI_CONFIG_PASSPHRASE=

sudo yum remove golang -y

GO_URL="https://go.dev/dl"
GO_FILE="go1.22.1.linux-amd64.tar.gz"

pushd /tmp &> /dev/null
curl -LO --progress-bar ${GO_URL}/${GO_FILE}
sudo tar -zxf ${GO_FILE} -C /usr/local
sudo rm -f ${GO_FILE}
popd &> /dev/null

export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/GO
export PATH=$PATH:$GOPATH/bin

go version

