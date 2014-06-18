/*
	Provides a simple command line interface to interact with a bitcoin rpc client to generate
	valid transactions in hex format
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/NSkelsey/btcbuilder"
	"github.com/conformal/btcwire"
)

var numouts *int = flag.Int("numout", 2, "The number of outpoints to create")
var valout *int = flag.Int("valout", 110000, "The value to set at each outpoint")

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ltime|log.Llongfile)
	logger.Println(valout, valout)

	params := btcbuilder.BuilderParams{
		Fee:      10000,
		DustAmnt: 546,
		InTarget: int64(*valout),
	}
	params = btcbuilder.SetParams(btcwire.TestNet3, params)

	list := []btcbuilder.TxBuilder{
		btcbuilder.NewSigHashSingleBuilder(params),
	}
	fanout := btcbuilder.NewFanOutBuilder(params, list, *numouts)

	fmt.Println(btcbuilder.Send(fanout, params))
}
