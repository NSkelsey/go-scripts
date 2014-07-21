/*
	Provides a simple command line interface to interact with a bitcoin rpc client to generate
	valid transactions in hex format
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/conformal/btcwire"
)

var cmd string = flag.String("cmd", "", "The type of tx to create. Your options are:\nsinglesighash\nopreturn\n")

func main() {
	flag.Parse()

	logger = log.New(os.Stdout, "", log.Ltime|log.Llongfile)
	client, currnet := setupNet(btcwire.TestNet3)
	defer client.Shutdown()

	var tx btcwire.MsgTx
	var err error
	switch cmd {
	case cmd == "sighashsingle":
		tx, err = buildSigHashSingle(client, currnet)
	case cmd == "opreturn":
		tx, err = buildDust(client, currnet)
	default:
		err = errors.New("No such type: %s", cmd)
	}
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("%s\n", toHex(tx))
}
