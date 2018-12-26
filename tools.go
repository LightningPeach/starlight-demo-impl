package main

import (
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// RoundSequenceNumber is defined as BaseSequenceNumber + RoundNumber * 4
func roundSequenceNumber(baseSequenceNumber, roundNumber int) int {
	return baseSequenceNumber + roundNumber * 4
}

func loadSequenceNumber(address string) (int, error) {
	account, err := horizon.DefaultTestNetClient.LoadAccount(address)
	if err != nil {
		return 0, err
	}
	sequenceNumber, err := strconv.Atoi(account.Sequence)
	if err != nil {
		return 0, err
	}
	return sequenceNumber, nil
}

func loadBalance(address string) string {
	account, err := horizon.DefaultTestNetClient.LoadAccount(address)
	if err != nil {
		log.Fatal(escrowAccount)
	}
	return account.Balances[0].Balance
}

func createAccount() (*keypair.Full, error) {
	pair, err := keypair.Random()
	if err != nil {
		return nil, err
	}

	//fmt.Println(pair.Seed())
	//fmt.Println(pair.Address())

	address := pair.Address()
	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + address)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	return pair, nil
}