package main

import (
	"github.com/conformal/btcscript"
	"github.com/conformal/btcwire"
)

type FanoutBuilder struct {
	Params   BuilderParams
	Builders []TxBuilder
	Copies   int64 // Number of copies to add
}

// A FanoutBuilder creates a transaction that has txouts set to the needed value
// for other tx builders that need those txouts as inputs
// The number of outputs created is len(builders)*copies + 1
func NewFanoutBuilder(params BuilderParams, builders []TxBuilder, copies int64) *FanoutBuilder {
	fb := FanoutBuilder{
		Params:   params,
		Builders: builders,
		Copies:   copies,
	}
	return &fb
}

func (fanB *FanoutBuilder) SatNeeded() int64 {
	sum := int64(0)
	for _, builder := range fanB.Builders {
		sum += builder.SatNeeded() * fanB.Copies
	}
	// Good Citizens pay the toll
	sum += fanB.Params.Fee
	return sum
}

func (fanB *FanoutBuilder) Build() (*btcwire.MsgTx, error) {
	totalSpent := fanB.SatNeeded()

	// Compose a set of Txins with enough to fund this transactions needs
	inParamSet, totalIn, err := composeUnspents(
		totalSpent,
		fanB.Params.Client,
		fanB.Params.NetParams)
	if err != nil {
		return nil, err
	}

	msgtx := btcwire.NewMsgTx()
	// funding inputs speced out with blank
	for _, inpParam := range inParamSet {
		txin := btcwire.NewTxIn(inpParam.OutPoint, []byte{})
		msgtx.AddTxIn(txin)
	}

	for i := range fanB.Builders {
		builder := fanB.Builders[i]
		amnt := builder.SatNeeded()
		for j := int64(0); j < fanB.Copies; j++ {
			addr, err := newAddr(fanB.Params.Client)
			if err != nil {
				return nil, err
			}
			script, _ := btcscript.PayToAddrScript(addr)
			txout := btcwire.NewTxOut(amnt, script)
			msgtx.AddTxOut(txout)
		}
	}

	changeAddr, err := newAddr(fanB.Params.Client)
	if err != nil {
		return nil, err
	}
	// change to solve unevenness
	change, ok := changeOutput(totalIn-totalSpent, fanB.Params.DustAmnt, changeAddr)
	if ok {
		msgtx.AddTxOut(change)
	}

	// sign msgtx for each input
	for i, inpParam := range inParamSet {
		privkey := inpParam.Wif.PrivKey.ToECDSA()
		subscript := inpParam.TxOut.PkScript
		scriptSig, err := btcscript.SignatureScript(msgtx, i, subscript,
			btcscript.SigHashAll, privkey, true)
		if err != nil {
			return nil, err
		}
		msgtx.TxIn[i].SignatureScript = scriptSig
	}

	return msgtx, nil
}
