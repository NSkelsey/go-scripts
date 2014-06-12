package main

import (
	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

type DustBuilder struct {
	// embeded consts
	Params BuilderParams
	// extending
	NumOuts int64
}

func NewDustBuilder(params BuilderParams, numOuts int64) (b *DustBuilder) {
	db := DustBuilder{
		Params:  params,
		NumOuts: numOuts,
	}
	return &db
}

func (builder *DustBuilder) SatNeeded() int64 {
	sum := builder.NumOuts*builder.Params.DustAmnt + builder.Params.Fee
	return sum
}

// A transaction that contains only dust ouputs and obeys the TxBuilder interface
func (builder *DustBuilder) Build() (*btcwire.MsgTx, error) {

	var inparams *TxInParams
	var err error
	inparams, err = specificUnspent(
		builder.SatNeeded(),
		builder.Params.Client,
		builder.Params.NetParams)
	if err != nil {
		return nil, err
	}

	oldTxOut := inparams.TxOut
	outpoint := inparams.OutPoint
	wifkey := inparams.Wif

	msgtx := btcwire.NewMsgTx()

	txin := btcwire.NewTxIn(outpoint, []byte{})
	msgtx.AddTxIn(txin)

	for i := int64(0); i < builder.NumOuts; i++ {
		addr, err := newAddr(builder.Params.Client)
		if err != nil {
			return nil, err
		}
		addrScript, err := btcscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		txOut := btcwire.NewTxOut(builder.Params.DustAmnt, addrScript)
		msgtx.AddTxOut(txOut)
	}

	// sign as usual
	privkey := wifkey.PrivKey.ToECDSA()
	sig, err := btcscript.SignatureScript(msgtx, 0, oldTxOut.PkScript, btcscript.SigHashAll, privkey, true)
	if err != nil {
		return nil, err
	}
	txin.SignatureScript = sig

	return msgtx, nil
}

func (b *DustBuilder) Log(s string) {
	b.Params.Logger.Printf(s)
}