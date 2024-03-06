#!/bin/bash
GO_URL="https://go.dev/dl"
GO_FILE="go1.22.1.linux-amd64.tar.gz"

cd /tmp
curl -LO --progress-bar ${GO_URL}/${GO_FILE}
sudo tar -zxf ${GO_FILE} -C /usr/local
sudo rm -f ${GO_FILE}

export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/GO
export PATH=$PATH:$GOPATH/bin

go version
