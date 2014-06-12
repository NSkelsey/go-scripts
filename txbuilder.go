package main

import (
	"fmt"
	"log"
	"os"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcwire"
)

type BuilderParams struct {
	Fee       int64
	DustAmnt  int64
	InTarget  int64 // The target input a transaction must be created with
	Logger    *log.Logger
	Client    *btcrpcclient.Client
	NetParams *btcnet.Params
}

type TxBuilder interface {
	// SatNeeded computes the specific value needed at an txout for the tx being built by the builder
	SatNeeded() int64
	// Build generates a MsgTx from the provided parameters, (rpc client, FEE, ...)
	Build() (*btcwire.MsgTx, error)
	// Log is short hand for logging in a tx builder with Param logger
	Log(string)
}

func CreateParams() BuilderParams {
	logger = log.New(os.Stdout, "", log.Ltime|log.Llongfile)
	client, params := setupNet(btcwire.TestNet3)

	bp := BuilderParams{
		Fee:       10000,
		DustAmnt:  546,
		InTarget:  100000,
		Logger:    logger,
		Client:    client,
		NetParams: params,
	}
	return bp
}

func main() {
	bp := CreateParams()
	dust := NewDustBuilder(bp, 10)

	jaun := make([]TxBuilder, 1)
	jaun[0] = dust
	fanout := NewFanoutBuilder(bp, jaun, 3)

	msg, err := fanout.Build()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(toHex(msg))
}
