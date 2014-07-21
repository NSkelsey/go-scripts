package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/NSkelsey/btcbuilder"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

var (
	client  *btcrpcclient.Client
	currnet *btcnet.Params
	params  btcbuilder.BuilderParams
	logger  *log.Logger
)

type AnonTxMessage struct {
	Tx string
}

type ProofMessage struct {
	Secret  string
	Address string
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
	_, err = btcutil.DecodeAddress(proof.Address, params.NetParams)
	if err != nil {
		logger.Printf("Bad address: %s\n", err)
		http.Error(w, "bad addr", 405)
		return
	}

	singleBuilder := btcbuilder.NewToAddrBuilder(params, proof.Address)
	// use builder interface
	// Generate the anonymous tx
	fundingtx, err := singleBuilder.Build()
	if err != nil {
		logger.Println(err)
		return
		http.Error(w, "Bad", 500)
	}

	logger.Println(btcbuilder.ToHex(fundingtx))
	message := AnonTxMessage{Tx: btcbuilder.ToHex(fundingtx)}
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
	bp := btcbuilder.BuilderParams{
		InTarget: 110000,
		Fee:      10000,
		DustAmnt: 546,
		Logger:   logger,
	}
	params = btcbuilder.SetParams(btcwire.TestNet3, bp)

	defer params.Client.Shutdown()

	http.HandleFunc("/", handler)
	where := "0.0.0.0:1050"
	logger.Printf("Listening on %s\n", where)
	err := http.ListenAndServeTLS(where, "cert.pem", "key.pem", nil)
	//err := http.ListenAndServe(where, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
