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
func selectUnspent(client *btcrpcclient.Client, minAmount int64) (btcwire.TxOut, btcwire.OutPoint, btcutil.WIF) {
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
