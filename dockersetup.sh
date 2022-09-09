#!/bin/bash
cd /3dps

#if [[ "$1" == "clean" ]]; then
#	rm -rf 3dps.go go.sum go.mod setup.sh wait-for-it.sh /bin/apk /usr/local/bin /usr/local/go /bin/busybox /bin/echo /bin/kill /bin/bash /usr/bin/printf
	# That would be hell to try to hack around, like how would you even start? Probably, find a file upload exploit would do it, besides you might as well just try to hack something more interesting than something no person would run
#	exec ./3dps
#fi


#openssl req -x509 -nodes -newkey rsa:2048 -keyout TLS.key -out TLS.crt -days 69420 -subj "/C=EA/ST=Planet/L=Earth/O=Global Security/OU=IT Department/CN=3DPS"
go mod tidy
go build 3dps.go
rm -rf 3dps.go setup.sh go.sum go.mod wait-for-it.sh /bin/apk /usr/local/bin /usr/local/go /bin/busybox /bin/echo /bin/kill /bin/bash /usr/bin/printf
exec ./3dps
# I was originally going to use postgres, but no one is going to get that many users
# If the sqlite experience is bad enough, I will try to use another database, ofc in another branch


#wget https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
#chmod +x wait-for-it.sh
#./wait-for-it.sh -h postgres -p 5432 -t 69 -- bash /3dps/setup.sh clean
