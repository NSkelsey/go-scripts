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
)

var (
	numouts *int    = flag.Int("numout", 2, "The number of outpoints to create")
	valout  *int    = flag.Int("valout", 110000, "The value to set at each outpoint")
	network *string = flag.String("network", "TestNet3", "The network to use")
	fee     *int    = flag.Int("fee", 50000, "The fee to pay for the transaction.")
)

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ltime|log.Llongfile)
	logger.Println(valout, valout)

	params := btcbuilder.BuilderParams{
		Fee:      int64(*fee),
		DustAmnt: 546,
		InTarget: int64(*valout),
	}

	btcparams, err := btcbuilder.NetParamsFromStr(*network)
	if err != nil {
		logger.Fatal(err)
	}
	params = btcbuilder.SetParams(btcparams.Net, params)

	single := btcbuilder.NewPayToPubKeyHash(params, 1)
	fmt.Println(single.Summarize())
	list := []btcbuilder.TxBuilder{
		single,
	}
	fanout := btcbuilder.NewFanOutBuilder(params, list, *numouts)

	fmt.Println(btcbuilder.Send(fanout, params))
}
