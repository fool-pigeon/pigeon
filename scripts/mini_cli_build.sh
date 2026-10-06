# Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

cd cmd/pg/ && CGO_ENABLE=0 go build -o $GOPATH/bin/pg -ldflags="-s -w"