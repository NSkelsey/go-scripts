package main

import (
	"bytes"
	"fmt"
)

func main() {
	bp := CreateParams()
	p2pkh := NewPayToPubKeyHash(bp, 2)
	dustBuilder := NewDustBuilder(bp, 3)
	sighash := NewSigHashSingleBuilder(bp)

	msg := "Find the cost of freedom buried in the ground. Mother earth will swallow you. Lay your body down. And who knows which is which and who is who. Help. Is there anybody out there."
	msgBytes := bytes.NewBufferString(msg).Bytes()
	key := newWifKeyPair(bp.NetParams)
	pklist := CreateList(msgBytes, key)
	//pklist := CreateList([]byte{}, key, key, key)
	multisig := NewMultiSigBuilder(bp, 1, pklist)

	jaun := []TxBuilder{p2pkh, dustBuilder, multisig, sighash}

	copies := int64(1)
	fanout := NewFanOutBuilder(bp, jaun, copies)

	fmt.Println("======= Run summary =======")
	fmt.Printf(fanout.Summarize())

	println("Create Fanout? If no create spend from [y/n]")
	g := "nope"
	fmt.Scanf("%s", &g)
	if g == "y" {
		resp := send(fanout, bp)
		fmt.Println("Sent fanout with txid: ", resp)
	}
	for i := int64(0); i < copies; i++ {
		resp := send(sighash, bp)
		//resp := send(multisig, bp)
		fmt.Println("Sent sighash single:\n", resp)
		//		resp := send(p2pkh, bp)
		//		fmt.Println(len(bp.PendingSet))
		//		fmt.Println("Sent p2pkh with txid:\n ", resp)
		//		resp = send(dustBuilder, bp)
		//		fmt.Println("Sent Dust with txid:\n", resp)
	}
}
