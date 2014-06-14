package main

import (
	"fmt"
	"log"
	"os"

	"github.com/conformal/btcjson"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcwire"
)

type BuilderParams struct {
	Fee        int64
	DustAmnt   int64
	InTarget   int64 // The target input a transaction must be created with
	Logger     *log.Logger
	Client     *btcrpcclient.Client
	NetParams  *btcnet.Params
	PendingSet map[string]struct{}
	List       []btcjson.ListUnspentResult
}

type TxBuilder interface {
	// SatNeeded computes the specific value needed at an txout for the tx being built by the builder
	SatNeeded() int64
	// Build generates a MsgTx from the provided parameters, (rpc client, FEE, ...)
	Build() (*btcwire.MsgTx, error)
	// Log is short hand for logging in a tx builder with Param logger
	Log(string)
	Summarize() string
}

func CreateParams() BuilderParams {
	logger = log.New(os.Stdout, "", log.Ltime|log.Llongfile)
	client, params := setupNet(btcwire.TestNet3)

	bp := BuilderParams{
		Fee:        10000,
		DustAmnt:   546,
		InTarget:   500000,
		Logger:     logger,
		Client:     client,
		NetParams:  params,
		PendingSet: make(map[string]struct{}),
		List:       make([]btcjson.ListUnspentResult, 0),
	}
	return bp
}

func main() {
	bp := CreateParams()
	//dust := NewDustBuilder(bp, 10)
	p2pkh := NewPayToPubKeyHash(bp, 2)
	dustBuilder := NewDustBuilder(bp, 3)

	jaun := make([]TxBuilder, 2)
	jaun[0] = p2pkh
	jaun[1] = dustBuilder

	copies := int64(4)
	fanout := NewFanOutBuilder(bp, jaun, copies)

	fmt.Println("======= Run summary =======")
	fmt.Printf(fanout.Summarize())

	//
	println("Proceed? [y/n]")
	g := "n"
	fmt.Scanf("%s", &g)
	if g == "y" {
		send(fanout, bp)
		//resp := send(fanout, bp)

		//		fmt.Println("Sent fanout with txid: ", resp)
		//		for i := int64(0); i < copies; i++ {
		//			resp = send(dustBuilder, bp)
		//			fmt.Println("Sent Dust with txid: ", resp)
		//			send(p2pkh, bp)
		//			fmt.Println("Sent p2pkh with txid: ", resp)
		//		}
	}
}

func send(builder TxBuilder, params BuilderParams) *btcwire.ShaHash {
	msg, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}
	println(toHex(msg))
	resp, err := params.Client.SendRawTransaction(msg, false)
	if err != nil {
		log.Fatal(err)
	}
	return resp
}
