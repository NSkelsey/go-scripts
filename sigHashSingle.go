package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

var currnet = btcnet.TestNet3Params
var pver = btcwire.ProtocolVersion
var magic = btcwire.TestNet3
var MinNeed = int64(5000)

func main() {

	connCfg := &btcrpcclient.ConnConfig{
		Host:         "localhost:18332",
		User:         "bitcoinrpc",
		Pass:         "EhxWGNKr1Z4LLqHtfwyQDemCRHF8gem843pnLj19K4go",
		HttpPostMode: true,
		DisableTLS:   true,
	}

	client, err := btcrpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// RPC to setup previous TX
	oldTxOut, outpoint, wifkey := selectUnspent(client, 5e3)

	// Transaction building

	txin := btcwire.NewTxIn(&outpoint, []byte{})

	// Change txout for us -- blank for the moment

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

	fmt.Printf("Tx: %s\n", toHex(msgtx))
}

func toHex(tx *btcwire.MsgTx) string {
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	tx.Serialize(buf)
	txHex := hex.EncodeToString(buf.Bytes())
	return txHex
}

func changeOutput(inVal, leave int64, addr btcutil.Address) *btcwire.TxOut {
	val := inVal - leave
	script, _ := btcscript.PayToAddrScript(addr)
	txout := btcwire.NewTxOut(val, script)
	return txout
}

func sumOutputs(tx *btcwire.MsgTx) (val int64) {
	val = 0
	for i := range tx.TxOut {
		val += tx.TxOut[i].Value
	}
	return val
}

func selectUnspent(client *btcrpcclient.Client, minAmount int64) (btcwire.TxOut, btcwire.OutPoint, btcutil.WIF) {
	list, err := client.ListUnspent()
	if err != nil {
		log.Fatal(err)
	}

	if len(list) < 1 {
		log.Fatal("Need more unspent txs")
	}
	var oldTxOut btcwire.TxOut
	var prevAddress btcutil.Address
	var outpoint btcwire.OutPoint
	for i := 0; i < len(list); i++ {
		prevJson := list[i]
		prevHash, _ := btcwire.NewShaHashFromStr(prevJson.TxId)
		vout := prevJson.Vout
		_prevTx, _ := client.GetRawTransaction(prevHash)
		prevTx := _prevTx.MsgTx()
		oldTxOut = *prevTx.TxOut[vout]
		prevAddress, _ = btcutil.DecodeAddress(prevJson.Address, &currnet)

		outpoint = *btcwire.NewOutPoint(prevHash, vout)
		if oldTxOut.Value >= minAmount {
			break
		}
	}

	if oldTxOut == new(btcwire.TxOut) {
		log.Fatal("Not enough BTC to fund tx here")
	}

	// Get private Key for address
	wifkey, _ := client.DumpPrivKey(prevAddress)
	return oldTxOut, outpoint, *wifkey
}

func wifToAddr(wifkey *btcutil.WIF) btcutil.Address {
	pubkey := wifkey.SerializePubKey()
	addr, _ := btcutil.NewAddressPubKeyHash(pubkey, &currnet)
	return addr
}
