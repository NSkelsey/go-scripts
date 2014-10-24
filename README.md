go-scripts
=========
[![Build Status](https://travis-ci.org/NSkelsey/go-scripts.svg?branch=master)](https://travis-ci.org/NSkelsey/go-scripts)

A collection of go-scripts that consume bitcoin available to the RPC user.


###anonfundserver

Generates a transaction which pays to the address provided by json that knows
the "secert."



###publish

Uses bitcoin available in bitcoind's wallet to generate a bulletin.



###fanout

Generates a set of transaction with outputs of a certain size. It is entirely 
configurable.
