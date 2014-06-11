package main

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

var pver = btcwire.ProtocolVersion

// Network specific config
var currnet btcnet.Params
var magic btcwire.BitcoinNet
var connCfg *btcrpcclient.ConnConfig

func pickNetwork(net btcwire.BitcoinNet) {
	var port string
	switch net {
	case btcwire.TestNet3:
		magic = btcwire.TestNet3
		currnet = btcnet.TestNet3Params
		port = "18332"
	case btcwire.MainNet:
		magic = btcwire.MainNet
		currnet = btcnet.MainNetParams
		port = "8332"
	case btcwire.SimNet:
		magic = btcwire.SimNet
		currnet = btcnet.SimNetParams
		port = "18554"
	}

	connCfg = &btcrpcclient.ConnConfig{
		Host:         "localhost:" + port,
		User:         "bitcoinrpc",
		Pass:         "EhxWGNKr1Z4LLqHtfwyQDemCRHF8gem843pnLj19K4go",
		HttpPostMode: true,
		DisableTLS:   true,
	}
}

func makeRpcClient() *btcrpcclient.Client {
	client, err := btcrpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	// check to see if we are connected
	client.GetBestBlock()
	return client
}

func specificUnspent(client *btcrpcclient.Client, targetAmnt int64) (btcwire.TxOut, btcwire.OutPoint, btcutil.WIF) {
	// gets an unspent output with an exact amount associated with it
	list, err := client.ListUnspent()
	if err != nil {
		log.Fatal(err)
	}

	var txOut *btcwire.TxOut
	var outPoint *btcwire.OutPoint
	var privKey *btcutil.WIF

	for i := 0; i < len(list); i++ {
		prevJson := list[i]
		inAmnt := toSatoshi(prevJson.Amount)
		if inAmnt == targetAmnt {
			// Found one, lets use it
			prevHash, _ := btcwire.NewShaHashFromStr(prevJson.TxId)
			outPoint = btcwire.NewOutPoint(prevHash, prevJson.Vout)
			script, _ := hex.DecodeString(prevJson.ScriptPubKey)
			txOut = btcwire.NewTxOut(inAmnt, script)

			prevAddress, _ := btcutil.DecodeAddress(prevJson.Address, &currnet)
			wifkey, _ := client.DumpPrivKey(prevAddress)
			return *txOut, *outPoint, *wifkey
		}
	}

	log.Fatalf("failed to a find an unspent with %s", targetAmnt)
	return *txOut, *outPoint, *privKey
}

func selectUnspent(client *btcrpcclient.Client, minAmount int64) (btcwire.TxOut, btcwire.OutPoint, btcutil.WIF) {
	// selects an unspent outpoint that is funded over the minAmount
	list, err := client.ListUnspent()
	if err != nil {
		log.Fatal(err)
	}

	if len(list) < 1 {
		log.Fatal("Need more unspent txs")
	}
	var oldTxOut *btcwire.TxOut
	var prevAddress btcutil.Address
	var outpoint btcwire.OutPoint
	for i := 0; i < len(list); i++ {
		prevJson := list[i]
		prevHash, _ := btcwire.NewShaHashFromStr(prevJson.TxId)
		vout := prevJson.Vout
		_prevTx, _ := client.GetRawTransaction(prevHash)
		prevTx := _prevTx.MsgTx()
		oldTxOut = prevTx.TxOut[vout]
		prevAddress, _ = btcutil.DecodeAddress(prevJson.Address, &currnet)

		outpoint = *btcwire.NewOutPoint(prevHash, vout)
		if oldTxOut.Value >= minAmount {
			break
		}
	}

	// Never found a good outpoint
	if oldTxOut == nil {
		log.Fatal("Not enough BTC to fund tx here")
	}

	// Get private Key for address
	wifkey, _ := client.DumpPrivKey(prevAddress)
	return *oldTxOut, outpoint, *wifkey
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
func wifToAddr(wifkey *btcutil.WIF) btcutil.Address {
	pubkey := wifkey.SerializePubKey()
	addr, _ := btcutil.NewAddressPubKeyHash(pubkey, &currnet)
	return addr
}

func newAddr(client *btcrpcclient.Client) btcutil.Address {
	addr, _ := client.GetNewAddress()
	return addr
}

func prevOutVal(client *btcrpcclient.Client, tx *btcwire.MsgTx) int64 {
	// requires an rpc client and outpoints within wallets realm
	total := int64(0)
	for i := range tx.TxIn {
		txin := tx.TxIn[i]
		prevTxHash := txin.PreviousOutpoint.Hash
		var tx *btcutil.Tx
		tx, err := client.GetRawTransaction(&prevTxHash)
		if err != nil {
			log.Fatalf("failed to find the tx, (its not in the wallet): %s\n", err)
		}
		vout := txin.PreviousOutpoint.Index
		txout := tx.MsgTx().TxOut[vout]
		total += txout.Value
	}
	return total
}

// Converts a float bitcoin into satoshi
func toSatoshi(m float64) int64 {
	return int64(float64(btcutil.SatoshiPerBitcoin) * m)
}
