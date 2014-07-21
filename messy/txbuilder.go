package main

import (
	"bytes"
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
		Fee:        50000,
		DustAmnt:   546,
		InTarget:   100000,
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
	p2pkh := NewPayToPubKeyHash(bp, 2)
	dustBuilder := NewDustBuilder(bp, 3)

	msg := "Find the cost of freedom buried in the ground. Mother earth will swallow you. Lay your body down. And who knows which is which and who is who. Help. Is there anybody out there."
	msgBytes := bytes.NewBufferString(msg).Bytes()
	key := newWifKeyPair(bp.NetParams)
	pklist := CreateList(msgBytes, key)
	//pklist := CreateList([]byte{}, key, key, key)
	multisig := NewMultiSigBuilder(bp, 1, pklist)

	jaun := []TxBuilder{p2pkh, dustBuilder, multisig}

	copies := int64(1)
	fanout := NewFanOutBuilder(bp, jaun, copies)

	fmt.Println("======= Run summary =======")
	fmt.Printf(fanout.Summarize())

	println("Create Fanout? If no create spend from [y/n]")
	g := "nope"
	fmt.Scanf("%s", &g)
	if g == "y" {
		resp := send(fanout, bp)
		fmt.Println("Sent fanout with txid: ", resp)
	}
	for i := int64(0); i < copies; i++ {
		resp := send(multisig, bp)
		fmt.Println("Sent multisig with txid:\n", resp)
		//		resp := send(p2pkh, bp)
		//		fmt.Println(len(bp.PendingSet))
		//		fmt.Println("Sent p2pkh with txid:\n ", resp)
		//		resp = send(dustBuilder, bp)
		//		fmt.Println("Sent Dust with txid:\n", resp)
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
