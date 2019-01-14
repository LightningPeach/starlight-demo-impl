package host

import (
	"fmt"
	"log"

	"github.com/LightningPeach/starlight-demo-impl/htlc_resolution"
	"github.com/LightningPeach/starlight-demo-impl/ratchet"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/LightningPeach/starlight-demo-impl/wire"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

type hostMessageCache struct {
	paymentProposeMsg     *wire.PaymentProposeMsg
	htlcPaymentProposeMsg *wire.HTLCPaymentProposeMsg
}

type Account struct {
	selfKeyPair *keypair.Full

	escrowKeyPair         *keypair.Full
	hostRatchetAccount    *ratchet.Account
	guestRatchetAccount   *ratchet.Account
	HtlcResolutionAccount *htlc_resolution.Account

	guestAddress string

	BaseSequenceNumber int

	cache *hostMessageCache
}

func New() (*Account, error) {
	fmt.Println("creating host account:")
	selfKeyPair, err := tools.CreateAccount()
	if err != nil {
		return nil, err
	}

	escrowKeyPair, err := keypair.Random()
	if err != nil {
		return nil, err
	}

	hostRatchetAccount, err := ratchet.New()
	if err != nil {
		return nil, err
	}

	guestRatchetAccount, err := ratchet.New()
	if err != nil {
		return nil, err
	}

	htlcResolutionAccount, err := htlc_resolution.New()
	if err != nil {
		return nil, err
	}

	hostAccount := &Account{
		selfKeyPair:           selfKeyPair,
		escrowKeyPair:         escrowKeyPair,
		hostRatchetAccount:    hostRatchetAccount,
		guestRatchetAccount:   guestRatchetAccount,
		HtlcResolutionAccount: htlcResolutionAccount,
		cache:                 new(hostMessageCache),
	}

	fmt.Printf("balance: %v\n\n", hostAccount.LoadBalance())
	return hostAccount, nil
}

func (host *Account) PublishTx(txe *build.TransactionEnvelopeBuilder) error {
	txeB64, err := txe.Base64()
	if err != nil {
		return err
	}

	// And finally, send it off to Stellar!
	if _, err := horizon.DefaultTestNetClient.SubmitTransaction(txeB64); err != nil {
		return err
	}

	return nil
}

func (host *Account) loadSequenceNumber() int {
	sequenceNumber, err := tools.LoadSequenceNumber(host.selfKeyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}

func (host *Account) LoadBalance() string {
	return tools.LoadBalance(host.selfKeyPair.Address())
}
