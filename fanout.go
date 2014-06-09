package main

import (
	"fmt"
	"log"

	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

func main() {
	tx := build(btcwire.TestNet3)
	fmt.Printf("%s\n", toHex(tx))
}

func build(net btcwire.BitcoinNet) *btcwire.MsgTx {
	pickNetwork(net)
	client := makeRpcClient()
	defer client.Shutdown()

	var spreadAmnt, fee, numOuts int64
	numOuts = 100
	spreadAmnt = 10000000
	fee = 20000
	// Total input must be spread Amnt + fee
	oldTxOut, outpoint, wifkey := selectUnspent(client, spreadAmnt+fee)

	msgtx := btcwire.NewMsgTx()
	// funding input
	txin := btcwire.NewTxIn(&outpoint, []byte{})
	msgtx.AddTxIn(txin)

	// fan out ouputs
	for i := int64(0); i < numOuts; i++ {
		addr := newAddr(client)
		script, _ := btcscript.PayToAddrScript(addr)
		amnt := spreadAmnt / numOuts
		txout := btcwire.NewTxOut(amnt, script)
		msgtx.AddTxOut(txout)
	}

	// change to solve unevenness
	change := changeOutput(oldTxOut.Value-sumOutputs(msgtx), fee, newAddr(client))
	msgtx.AddTxOut(change)

	// sign msgtx
	privkey := wifkey.PrivKey.ToECDSA()
	scriptSig, err := btcscript.SignatureScript(msgtx, 0, oldTxOut.PkScript,
		btcscript.SigHashAll, privkey, true)
	if err != nil {
		log.Fatal("Signing Failed", err)
	}
	txin.SignatureScript = scriptSig

	return msgtx
}
