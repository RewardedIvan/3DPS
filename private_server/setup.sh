#!/bin/bash
cd /3dps
#openssl req -x509 -nodes -newkey rsa:2048 -keyout TLS.key -out TLS.crt -days 69420 -subj "/C=EA/ST=Planet/L=Earth/O=Global Security/OU=IT Department/CN=3DPS"
go run 3dps.go
go mod init github.com/RewardedIvan/3DPS
go mod tidy
go install github.com/mattn/go-sqlite3@latest
go run 3dps.go
