#/bin/bash
curl --verbose --cacert cert.pem -H "Content-Type: application/json" -d '{"secret":"bilbo has the ring", "Address":"n3wDLcEM3mKoiBp9YBbfngXS2e9s1GiBWx"}' https://127.0.0.1:1050/
