package main

import (
	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

var forUse int64 = 100000
var FEE int64 = 10000

func buildSigHashSingle(client *btcrpcclient.Client, net *btcnet.Params) (*btcwire.MsgTx, error) {
	// RPC to setup previous TX
	txInParams, err := selectUnspent(forUse+FEE, client, net)
	if err != nil {
		return nil, err
	}

	oldTxOut := txInParams.TxOut
	outpoint := txInParams.OutPoint
	wifkey := txInParams.Wif

	// Transaction building

	txin := btcwire.NewTxIn(outpoint, []byte{})

	// notice amount in
	total := oldTxOut.Value
	change := changeOutput(total-(forUse+FEE), wifToAddr(wifkey, net))
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
		logger.Fatal("ScriptSig failed", err)
	}

	txin.SignatureScript = scriptSig
	// This demonstrates that we can sign and then permute a txout
	//msgtx.TxOut[1].PkScript = oldTxOut.PkScript
	blank.Value = forUse

	return msgtx, nil
}
