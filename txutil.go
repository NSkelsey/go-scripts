package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

var pver = btcwire.ProtocolVersion
var logger *log.Logger

// Network specific config
var magic btcwire.BitcoinNet

// Everything you need to spend from a txout in the UTXO
type TxInParams struct {
	TxOut    *btcwire.TxOut
	OutPoint *btcwire.OutPoint
	Wif      *btcutil.WIF
}

func pickNetwork(net btcwire.BitcoinNet) (btcrpcclient.ConnConfig, *btcnet.Params) {
	var currnet btcnet.Params
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
	case btcwire.TestNet:
		magic = btcwire.TestNet
		currnet = btcnet.RegressionNetParams
		port = "18443"
	}

	connCfg := btcrpcclient.ConnConfig{
		Host:         "localhost:" + port,
		User:         "bitcoinrpc",
		Pass:         "EhxWGNKr1Z4LLqHtfwyQDemCRHF8gem843pnLj19K4go",
		HttpPostMode: true,
		DisableTLS:   true,
	}
	return connCfg, &currnet
}

// Sets up an RPC client configured for the selected network,
// it also responds with the relevant btcnet.Params struct
func setupNet(net btcwire.BitcoinNet) (*btcrpcclient.Client, *btcnet.Params) {
	connCfg, netparams := pickNetwork(net)
	client := makeRpcClient(connCfg)
	return client, netparams
}

func makeRpcClient(connCfg btcrpcclient.ConnConfig) *btcrpcclient.Client {
	client, err := btcrpcclient.New(&connCfg, nil)
	if err != nil {
		logger.Fatal(err)
	}
	// check to see if we are connected
	_, err = client.GetDifficulty()
	if err != nil {
		logger.Fatal(err)
	}
	return client
}

// specificUnspent gets an unspent output with an exact amount associated with it.
// it throws an error otherwise
func specificUnspent(targetAmnt int64, client *btcrpcclient.Client, net *btcnet.Params) (*TxInParams, error) {
	list, err := client.ListUnspent()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(list); i++ {
		prevJson := list[i]
		inAmnt := toSatoshi(prevJson.Amount)
		if inAmnt == targetAmnt {
			// Found one, lets use it
			// None of these ~should~ ever throw errors
			prevHash, _ := btcwire.NewShaHashFromStr(prevJson.TxId)
			outPoint := btcwire.NewOutPoint(prevHash, prevJson.Vout)
			script, _ := hex.DecodeString(prevJson.ScriptPubKey)
			txOut := btcwire.NewTxOut(inAmnt, script)

			prevAddress, _ := btcutil.DecodeAddress(prevJson.Address, net)
			wifkey, err := client.DumpPrivKey(prevAddress)
			if err != nil {
				return nil, err
			}
			inParams := TxInParams{
				TxOut:    txOut,
				OutPoint: outPoint,
				Wif:      wifkey,
			}
			return &inParams, nil
		}
	}
	return nil, errors.New("Could not find a txout with specific amount.")
}

// selectUnspent picks an unspent output that has atleast minAmount (sats) associated with it.
// It throws an error otherwise
func selectUnspent(minAmount int64, client *btcrpcclient.Client, net *btcnet.Params) (*TxInParams, error) {
	// selects an unspent outpoint that is funded over the minAmount
	list, err := client.ListUnspent()
	if err != nil {
		logger.Println("list unpsent threw")
		return nil, err
	}

	if len(list) < 1 {
		return nil, errors.New("No unspent outputs at all.")
	}

	for i := 0; i < len(list); i++ {
		prevJson := list[i]
		inAmnt := toSatoshi(prevJson.Amount)
		if inAmnt >= minAmount {
			// Found one, lets use it
			// None of these ~should~ ever throw errors
			prevHash, _ := btcwire.NewShaHashFromStr(prevJson.TxId)
			outPoint := btcwire.NewOutPoint(prevHash, prevJson.Vout)
			script, _ := hex.DecodeString(prevJson.ScriptPubKey)
			txOut := btcwire.NewTxOut(inAmnt, script)

			prevAddress, _ := btcutil.DecodeAddress(prevJson.Address, net)
			wifkey, err := client.DumpPrivKey(prevAddress)
			if err != nil {
				return nil, err
			}
			inParams := TxInParams{
				TxOut:    txOut,
				OutPoint: outPoint,
				Wif:      wifkey,
			}
			return &inParams, nil
		}
	}
	// Never found a good outpoint
	return nil, errors.New("No txout with enough funds")
}

// toHex converts a msgTx into a hex string.
func toHex(tx *btcwire.MsgTx) string {
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	tx.Serialize(buf)
	txHex := hex.EncodeToString(buf.Bytes())
	return txHex
}

// generates a change output funding provided addr
func changeOutput(change int64, addr btcutil.Address) *btcwire.TxOut {
	script, err := btcscript.PayToAddrScript(addr)
	if err != nil {
		log.Fatalf("failed to create script: %s\n", err)
	}
	txout := btcwire.NewTxOut(change, script)
	return txout
}

// sumOutputs derives the values in satoshis of tx.
func sumOutputs(tx *btcwire.MsgTx) (val int64) {
	val = 0
	for i := range tx.TxOut {
		val += tx.TxOut[i].Value
	}
	return val
}

func wifToAddr(wifkey *btcutil.WIF, net *btcnet.Params) btcutil.Address {
	pubkey := wifkey.SerializePubKey()
	pkHash := btcutil.Hash160(pubkey)
	addr, err := btcutil.NewAddressPubKeyHash(pkHash, net)
	if err != nil {
		log.Fatalf("failed to convert wif to address: %s\n", err)
	}
	return addr
}

// Gets a new address from an rpc client, catches all errors
func newAddr(client *btcrpcclient.Client) btcutil.Address {
	addr, err := client.GetNewAddress()
	if err != nil {
		log.Fatal(err)
	}
	return addr
}

// prevOutVal looks up all the values of the oupoints used in the current tx
func prevOutVal(tx *btcwire.MsgTx, client *btcrpcclient.Client) (int64, error) {
	// requires an rpc client and outpoints within wallets realm
	total := int64(0)
	for i := range tx.TxIn {
		txin := tx.TxIn[i]
		prevTxHash := txin.PreviousOutpoint.Hash
		var tx *btcutil.Tx
		tx, err := client.GetRawTransaction(&prevTxHash)
		if err != nil {
			return -1, err
		}
		vout := txin.PreviousOutpoint.Index
		txout := tx.MsgTx().TxOut[vout]
		total += txout.Value
	}
	return total, nil
}

// Converts a float bitcoin into satoshi
func toSatoshi(m float64) int64 {
	return int64(float64(btcutil.SatoshiPerBitcoin) * m)
}
