package main

import (
	"log"

	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

var minNeeded = int64(5000)

func buildSigHashSingle() *btcwire.MsgTx {

	pickNetwork(btcwire.TestNet3)
	client := makeRpcClient()
	defer client.Shutdown()

	// RPC to setup previous TX
	oldTxOut, outpoint, wifkey := selectUnspent(client, minNeeded)

	// Transaction building

	txin := btcwire.NewTxIn(&outpoint, []byte{})

	// Change txout for us -- blank for the moment

	// notice amount in
	change := changeOutput(oldTxOut.Value, 2555, wifToAddr(&wifkey))
	// Blank permutable txout for users to play with
	blank := btcwire.NewTxOut(0, []byte{})

	msgtx := btcwire.NewMsgTx()
	msgtx.AddTxIn(txin)
	msgtx.AddTxOut(change)
	msgtx.AddTxOut(blank)

	subscript := oldTxOut.PkScript
	privkey := wifkey.PrivKey.ToECDSA()
	scriptSig, err := btcscript.SignatureScript(msgtx, 0, subscript, btcscript.SigHashSingle, privkey, true)
	if err != nil {
		log.Fatal("ScriptSig failed", err)
	}

	txin.SignatureScript = scriptSig
	// This demonstrates that we can sign and then permute a txout
	blank.PkScript = oldTxOut.PkScript
	blank.Value = 556

	return msgtx
}

//func main() {
//	tx := buildSigHashSingle()
//	fmt.Printf("%s\n", toHex(tx))
//}
