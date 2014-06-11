#/bin/bash
curl --verbose --cacert cert.pem -H "Content-Type: application/json" -d '{"secret":"bilbo has the ring"}' https://ahimsa.io:1050/
