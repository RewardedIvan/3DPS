#!/bin/sh
cd /src

# Check the wiki for more info
#openssl req -x509 -nodes -newkey rsa:2048 -keyout TLS.key -out TLS.crt -days 69420 -subj "/C=EA/ST=Planet/L=Earth/O=Global Security/OU=IT Department/CN=3DPS"

go mod tidy
go build
mv 3DPS /3DPS/server
cd /
rm -rf /src

# I was originally going to use postgres, but no one is going to get that many users
# If the sqlite experience is bad enough, I will try to use another database, ofc in another branch

#wget https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
#chmod +x wait-for-it.sh
#./wait-for-it.sh -h postgres -p 5432 -t 69 -- ......
