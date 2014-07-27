// Builds and sends a bulletin from command line parameters. This script
// automatically reads your .bitcoin directory to pull out the RPC parameters
// and sends from some arbitrary address supplied in your wallet.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/NSkelsey/btcbuilder"
	"github.com/NSkelsey/btcsubprotos/ahimsa"
)

var (
	topic    *string = flag.String("topic", "Let It Be Known", "The topic of the bulletin")
	msg      *string = flag.String("msg", "", "The message to send")
	network  *string = flag.String("network", "TestNet3", "The network to use")
	fee      *int    = flag.Int("fee", 50000, "Fee in satoshi to pay miners")
	burnAmnt *int    = flag.Int("burn", 10000, "Amount of satoshi to burn for each txout")
)

func main() {
	flag.Parse()

	params := btcbuilder.BuilderParams{
		Fee:      int64(*fee),
		DustAmnt: 546,
	}

	btcparams, err := btcbuilder.NetParamsFromStr(*network)
	if err != nil {
		log.Fatal(err)
	}
	params = btcbuilder.SetParams(btcparams.Net, params)

	bltn := ahimsa.Bulletin{
		Topic:   *topic,
		Message: *msg,
	}

	builder := btcbuilder.NewBulletinBuilder(params, int64(*burnAmnt), bltn)

	fmt.Println(builder.Summarize())
	fmt.Println(btcbuilder.Send(builder, params))
}
