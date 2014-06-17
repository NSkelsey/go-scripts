package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

type NullDataBuilder struct {
	Params BuilderParms
	Data   []byte
	Change bool
}

func NewNullData(params BuilderParams, data []byte, change bool) {
	ndB := NewNullData{
		Params: params,
		Data:   data,
		Change: change,
	}
	return &ndB
}

func (nbD *NullDataBuilder) SatNeeded() (sum int64) {
	sum = 0
	if change {
		sum = nbD.Params.InTarget
	} else {
		sum = nbD.Params.DustAmnt + nbD.Params.Fee
	}
	return sum
}

func (nbD *NullDataBuilder) Build() (*btcwire.Msgtx, error) {

	utxo := selectSpecificUnspent(ndB.SatNeeded(), nbD.Params)

	msgtx := btcwire.NewMsgTx()

	if len(nbD.Data) > 40 {
		return nil, errors.New("Data is too long to make this a standard tx.")
	}

	// OP Return output
	retbuilder := btcscript.NewScriptBuilder().AddOp(btcscript.OP_RETURN).AddData(ndB.Data)
	op_return := btcwire.NewTxOut(0, retbuilder.Script())
	msgtx.AddTxOut(op_return)

	if ndB.Change {
		// change ouput
		change := changeOutput(nbD.SatNeeded()-nbD.Params.Fee, newAddr(client))
		msgtx.AddTxOut(change)
	}

	// funding input
	txin := btcwire.NewTxIn(utxo.OutPoint, []byte{})
	msgtx.AddTxIn(txin)

	// sign msgtx
	privkey := utxo.Wif.PrivKey.ToECDSA()
	scriptSig, err := btcscript.SignatureScript(msgtx, 0, utxo.TxOut.PkScript, btcscript.SigHashAll, privkey, true)
	if err != nil {
		log.Fatal("Signing Failed", err)
	}
	txin.SignatureScript = scriptSig

	return msgtx
}

func (nbD *NullDataBuilder) Log(msg string) {
	nbD.Params.Logger.Println(msg)
}

func (nbD *NullDataBuilder) Summarize() string {
	s := "==== NullData ====\nSatNeeded:\t%d\nTxIns:\t1\nTxOuts:\t%d\nLenData:\t%d\n"
	numouts := 1
	if change {
		numouts = 2
	}
	return fmt.Sprintf(s, nbD.SatNeeded(), numouts, len(nbD.Data))
}
