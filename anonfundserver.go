package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcwire"
)

var client *btcrpcclient.Client
var currnet *btcnet.Params

type AnonTxMessage struct {
	Tx    string
	Value int64
}

type ProofMessage struct {
	Secret string
}

func handler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Request!: %s\n", req.Method)
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if 0 > req.ContentLength || req.ContentLength > 1024 {
		http.Error(w, "GTFO", 411)
		return
	}
	buf := make([]byte, req.ContentLength)
	_, err := req.Body.Read(buf)
	if err != nil {
		log.Println("failed to read body: %s\n", err)
		http.Error(w, "bad buffer", 500)
		return
	}

	logger.Printf("POST: %s\n", buf)

	var proof ProofMessage
	if err := json.Unmarshal(buf, &proof); err != nil {
		logger.Printf("failed to parse json: %s\n", err)
		http.Error(w, "bad json", 500)
		return
	}

	if !check(proof) {
		logger.Printf("Did not pass test\n")
		http.Error(w, "bad proof", 405)
		return
	}
	// Generate the anonymous tx
	fundingtx, err := buildSigHashSingle(client, currnet)
	if err != nil {
		logger.Println(err)
		http.Error(w, "Bad", 500)
	}
	prevVal, err := prevOutVal(fundingtx, client)
	if err != nil {
		logger.Println(err)
		http.Error(w, "Bad", 500)
	}
	message := AnonTxMessage{Tx: toHex(fundingtx), Value: prevVal}
	bytes, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Cannot serialize the tx", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	logger.Println("Message sent")
}

func check(proof ProofMessage) bool {
	// checks if the sender has suffecient proof to get some coins
	return proof.Secret == "bilbo has the ring"
}

func main() {

	logger = log.New(os.Stdout, "", log.Ltime)
	client, currnet = setupNet(btcwire.TestNet3)
	defer client.Shutdown()

	http.HandleFunc("/", handler)
	where := "0.0.0.0:1050"
	logger.Printf("Listening on %s\n", where)
	err := http.ListenAndServeTLS(where, "cert.pem", "key.pem", nil)
	//err := http.ListenAndServe(where, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
