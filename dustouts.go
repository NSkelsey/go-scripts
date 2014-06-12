package main

import (
	"fmt"
	"log"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

var logger log.Logger
var FEE int64 = 10000
var numOuts int64 = 15
var dustAmnt, targetOut int64 = 546, 546*numOuts + FEE

type 

func buildDust(client *btcrpcclient.Client, net *btcnet.Params) (*btcwire.MsgTx, error) {
	// A transaction that contains only dust ouputs and pays the regular fees
	pickNetwork(btcwire.TestNet3)
	client := makeRpcClient()
	defer client.Shutdown()

	// will die if it cannot find a tx
	oldTxOut, outpoint, wifkey := specificUnspent(client, targetOut)

	msgtx := btcwire.NewMsgTx()

	txin := btcwire.NewTxIn(&outpoint, []byte{})
	msgtx.AddTxIn(txin)

	for i := int64(0); i < numOuts; i++ {
		addr := newAddr(client)
		addrScript, err := btcscript.PayToAddrScript(addr)
		if err != nil {
			log.Fatalf("failed to create script from addr: %s\n", err)
		}
		txOut := btcwire.NewTxOut(dustAmnt, addrScript)
		msgtx.AddTxOut(txOut)
	}

	// sign as usual
	privkey := wifkey.PrivKey.ToECDSA()
	sig, err := btcscript.SignatureScript(msgtx, 0, oldTxOut.PkScript, btcscript.SigHashAll, privkey, true)
	if err != nil {
		log.Fatalf("failed to sign msgtx: %s\n", err)
	}
	txin.SignatureScript = sig

	return msgtx
}

func main() {
	tx := buildDust()
	fmt.Printf("%s\n", toHex(tx))
}
