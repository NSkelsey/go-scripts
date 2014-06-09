package main

import (
	"fmt"
	"log"

	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

func build(net btcwire.BitcoinNet) *btcwire.MsgTx {
	pickNetwork(btcwire.TestNet3)
	client := makeRpcClient()
	defer client.Shutdown()

	oldTxOut, outpoint, wifkey := selectUnspent(client, 2000)

	msgtx := btcwire.NewMsgTx()

	// OP Return output
	derp := []byte("This is an attempt to see what works BLA")
	retbuilder := btcscript.NewScriptBuilder().AddOp(btcscript.OP_RETURN).AddData(derp)
	op_return := btcwire.NewTxOut(0, retbuilder.Script())
	msgtx.AddTxOut(op_return)

	// change ouput
	change := changeOutput(oldTxOut.Value, 1000, newAddr(client))
	msgtx.AddTxOut(change)

	// funding input
	txin := btcwire.NewTxIn(&outpoint, []byte{})
	msgtx.AddTxIn(txin)

	// sign msgtx
	privkey := wifkey.PrivKey.ToECDSA()
	scriptSig, err := btcscript.SignatureScript(msgtx, 0, oldTxOut.PkScript, btcscript.SigHashAll, privkey, true)
	if err != nil {
		log.Fatal("Signing Failed", err)
	}
	txin.SignatureScript = scriptSig

	return msgtx
}

func main() {
	msgtx := build(TestNet3)
	fmt.Printf("%s\n", toHex(msgtx))
}
