#/bin/bash
curl --cacert cert.pem -H "Content-Type: application/json" -d '{"secret":"bilbo has the ring"}' https://127.0.0.1:1050/single
