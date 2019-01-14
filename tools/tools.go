package tools

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

const (
	DefaultMaxRoundDuration = 10
	DefaultFinalityDelay    = 10
	DefaultFeerate          = "undefined"
	DefaultHostAmount       = "500"
	DefaultPaymentAmount    = "100"

	baseFee = 0.00001
	feeRate = baseFee
)

var HtlcMode bool

func GetBlockChainTime() uint64 {
	return uint64(time.Now().Unix())
}

type AccountType uint8

const (
	HostRatchetAccount AccountType = iota
	GuestRatchetAccount
	HtlcResolutionAccountType
	EscrowAccount
)

func (account AccountType) String() string {
	switch account {
	case HostRatchetAccount:
		return "<host_ratchet_account>"
	case GuestRatchetAccount:
		return "<guest_ratchet_account>"
	case HtlcResolutionAccountType:
		return "<htlc_resolution_account_type>"
	case EscrowAccount:
		return "<escrow_account>"
	default:
		return "<unknown>"
	}
}

type State uint8

const (
	FundingState State = iota
	PaymentState
	HtlcPaymentState
)

// RoundSequenceNumber is defined as BaseSequenceNumber + RoundNumber * 4
func RoundSequenceNumber(baseSequenceNumber, roundNumber int) int {
	return baseSequenceNumber + roundNumber*4
}

func LoadSequenceNumber(address string) (int, error) {
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

func LoadBalance(address string) string {
	account, err := horizon.DefaultTestNetClient.LoadAccount(address)
	if err != nil {
		log.Fatal(err)
	}
	return account.Balances[0].Balance
}

func CreateAccount() (*keypair.Full, error) {
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

func ShowDetailError(err error) {
	fmt.Println("detail error description:")

	err2 := err.(*horizon.Error).Problem
	tmpl := `
	Type:     %v
	Title:    %v
	Status:   %v
	Detail:   %v
	Instance: %v
	`
	fmt.Printf(tmpl, err2.Type, err2.Title, err2.Status, err2.Detail, err2.Instance)

	for key, value := range err2.Extras {
		fmt.Printf("key: %v, value: %v\n", key, string(value))
	}
}
