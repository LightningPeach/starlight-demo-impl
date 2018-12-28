package main

import (
	"fmt"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"
	"log"
	"time"
)

const (
	defaultMaxRoundDuration = 10
	defaultFinalityDelay    = 10
	defaultFeerate          = "undefined"
	defaultHostAmount       = "500"
	defaultPaymentAmount    = "100"

	baseFee = 0.00001
	feeRate = baseFee
)

func getBlockChainTime() uint64 {
	return uint64(time.Now().Unix())
}

type accountType uint8

const (
	hostRatchetAccount accountType = iota
	guestRatchetAccount
	escrowAccount
)

func (account accountType) String() string {
	switch account {
	case hostRatchetAccount:
		return "<host_ratchet_account>"
	case guestRatchetAccount:
		return "<guest_ratchet_account>"
	case escrowAccount:
		return "<escrow_account>"
	default:
		return "<unknown>"
	}
}

type hostMessageCache struct {
	paymentProposeMsg *PaymentProposeMsg
	// paymentAcceptMsg *PaymentAcceptMsg
}

type hostAccount struct {
	selfKeyPair *keypair.Full

	escrowKeyPair       *keypair.Full
	hostRatchetAccount  *ratchetAccount
	guestRatchetAccount *ratchetAccount

	guestAddress string

	baseSequenceNumber int

	cache *hostMessageCache
}

func newHostAccount() (*hostAccount, error) {
	fmt.Println("creating host account:")
	selfKeyPair, err := createAccount()
	if err != nil {
		return nil, err
	}

	escrowKeyPair, err := keypair.Random()
	if err != nil {
		return nil, err
	}

	hostRatchetAccount, err := newRatchetAccount()
	if err != nil {
		return nil, err
	}

	guestRatchetAccount, err := newRatchetAccount()
	if err != nil {
		return nil, err
	}

	hostAccount := &hostAccount{
		selfKeyPair:         selfKeyPair,
		escrowKeyPair:       escrowKeyPair,
		hostRatchetAccount:  hostRatchetAccount,
		guestRatchetAccount: guestRatchetAccount,
		cache:               new(hostMessageCache),
	}

	fmt.Printf("balance: %v\n\n", hostAccount.loadBalance())
	return hostAccount, nil
}

func (host *hostAccount) setupAccountTx(account accountType) error {
	fmt.Printf("creating: %v\n", account)
	var dest string
	switch account {
	case hostRatchetAccount:
		dest = host.hostRatchetAccount.keyPair.Address()
	case guestRatchetAccount:
		dest = host.guestRatchetAccount.keyPair.Address()
	case escrowAccount:
		dest = host.escrowKeyPair.Address()
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.selfKeyPair.Address()},
		build.Sequence{Sequence: uint64(host.loadSequenceNumber()) + 1},
		build.CreateAccount(
			build.Destination{AddressOrSeed: dest},
			build.NativeAmount{Amount: "1"},
		),
	)
	if err != nil {
		return err
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(host.selfKeyPair.Seed())
	if err != nil {
		return err
	}

	if err := host.publishTx(&txe); err != nil {
		return err
	}

	if account == escrowAccount {
		sequence, err := loadSequenceNumber(host.escrowKeyPair.Address())
		if err != nil {
			return err
		}
		host.baseSequenceNumber = sequence
	}
	fmt.Printf("balance: %v\n\n", loadBalance(dest))
	return nil
}

func (host *hostAccount) publishFundingTx(guestEscrowPubKey string) error {
	fmt.Println("publish funding tx")
	host.guestAddress = guestEscrowPubKey
	fundingTime := getBlockChainTime()

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.selfKeyPair.Address()},
		build.Sequence{Sequence: uint64(host.loadSequenceNumber()) + 1},
		build.Timebounds{
			MaxTime: uint64(fundingTime + defaultMaxRoundDuration + defaultFinalityDelay),
		},
		// Escrow Account
		build.Payment(
			build.Destination{AddressOrSeed: host.escrowKeyPair.Address()},
			// TODO(evg): refactor it
			build.NativeAmount{Amount: "500.50008"}, // defaultHostAmount + 0.5 + 8 * feerate
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.escrowKeyPair.Address()},
			build.SetThresholds(2, 2, 2),
			build.Signer{
				Address: guestEscrowPubKey,
				Weight:  1,
			},
		),
		// GuestRatchetAccount
		build.Payment(
			build.Destination{AddressOrSeed: host.guestRatchetAccount.keyPair.Address()},
			build.NativeAmount{Amount: "1.00001"}, // 1 + 1 * feerate
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.guestRatchetAccount.keyPair.Address()},
			build.MasterWeight(0),
			build.SetThresholds(2, 2, 2),
			build.Signer{
				Address: guestEscrowPubKey,
				Weight:  1,
			},
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.guestRatchetAccount.keyPair.Address()},
			build.Signer{
				Address: host.escrowKeyPair.Address(),
				Weight:  1,
			},
		),
		// HostRatchetAccount
		build.Payment(
			build.Destination{AddressOrSeed: host.hostRatchetAccount.keyPair.Address()},
			build.NativeAmount{Amount: "0.50001"}, // 0.5 + 1 * feerate
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.hostRatchetAccount.keyPair.Address()},
			build.MasterWeight(0),
			build.Signer{
				Address: host.escrowKeyPair.Address(),
				Weight:  1,
			},
		),
	)
	if err != nil {
		return err
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(
		host.selfKeyPair.Seed(),
		host.escrowKeyPair.Seed(),
		host.hostRatchetAccount.keyPair.Seed(),
		host.guestRatchetAccount.keyPair.Seed(),
	)
	if err != nil {
		return err
	}

	return host.publishTx(&txe)
}

func (host *hostAccount) cleanupTx() {}

func (host *hostAccount) publishRatchetTx(sig *xdr.DecoratedSignature) error {
	fmt.Println("publish ratchet tx")
	tx, err := host.createAndSignRatchetTxForSelf(
		sig,
		host.cache.paymentProposeMsg.PaymentTime,
		roundSequenceNumber(host.baseSequenceNumber, host.cache.paymentProposeMsg.RoundNumber),
	)
	if err != nil {
		return err
	}
	if err := host.publishTx(tx); err != nil {
		showDetailError(err)
		return err
	}
	return nil
}

func (host *hostAccount) createAndSignRatchetTxForSelf(
	sig *xdr.DecoratedSignature,
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := host.createRatchetTxForSelf(paymentTime, roundSequenceNumber)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return nil, err
	}

	txe.E.Signatures = append(txe.E.Signatures, *sig)
	return &txe, nil
}

func (host *hostAccount) createRatchetTxForSelf(
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionBuilder,
	error,
) {
	return createRatchetTx(
		host.hostRatchetAccount.keyPair.Address(),
		host.escrowKeyPair.Address(),
		paymentTime,
		roundSequenceNumber,
	)
}

func (host *hostAccount) settleOnlyWithHostTx(
	guestSettleOnlyWithHostSig *xdr.DecoratedSignature,
	fundingTime uint64,
) error {
	rsn := roundSequenceNumber(host.baseSequenceNumber, 1)

	settleOnlyWithHostTx, err := createSettleOnlyWithHostTx(
		host.selfKeyPair.Address(),
		host.escrowKeyPair.Address(),
		host.guestRatchetAccount.keyPair.Address(),
		host.hostRatchetAccount.keyPair.Address(),
		fundingTime,
		rsn,
	)
	if err != nil {
		return err
	}

	txe, err := settleOnlyWithHostTx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return err
	}

	txe.E.Signatures = append(txe.E.Signatures, *guestSettleOnlyWithHostSig)

	if err := host.publishTx(&txe); err != nil {
		showDetailError(err)
		return err
	}
	return nil
}

func (host *hostAccount) publishSettleWithGuestTx(
	rsn,
	paymentTime uint64,
	guestAmount string,
	sig *xdr.DecoratedSignature,
) error {
	fmt.Println("publish settle with guest tx")

	tx, err := host.createAndSignSettleWithGuestTx(rsn, paymentTime, guestAmount)
	if err != nil {
		return err
	}
	tx.E.Signatures = append(tx.E.Signatures, *sig)

	if err := host.publishTx(tx); err != nil {
		showDetailError(err)
		return err
	}
	return nil
}

func (host *hostAccount) createAndSignSettleWithGuestTx(
	rsn uint64,
	paymentTime uint64,
	guestAmount string,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := createSettleWithGuestTx(
		rsn,
		paymentTime,
		host.guestAddress,
		guestAmount,
		host.escrowKeyPair.Address(),
	)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}

func (host *hostAccount) publishSignSettleWithHostTx(rsn, paymentTime uint64, sig *xdr.DecoratedSignature) error {
	fmt.Println("publish settle with host tx")

	tx, err := host.createAndSignSettleWithHostTx(rsn, paymentTime)
	if err != nil {
		return err
	}
	tx.E.Signatures = append(tx.E.Signatures, *sig)

	if err := host.publishTx(tx); err != nil {
		showDetailError(err)
		return err
	}
	return nil
}

func (host *hostAccount) createAndSignSettleWithHostTx(rsn, paymentTime uint64) (*build.TransactionEnvelopeBuilder, error) {
	tx, err := createSettleWithHostTx(
		rsn,
		paymentTime,
		host.escrowKeyPair.Address(),
		host.guestRatchetAccount.keyPair.Address(),
		host.hostRatchetAccount.keyPair.Address(),
		host.selfKeyPair.Address(),
	)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return nil, err
	}
	return &txe, nil
}

func (host *hostAccount) publishTx(txe *build.TransactionEnvelopeBuilder) error {
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

func (host *hostAccount) createChannelProposeMsg(guestEscrowPubKey string) *ChannelProposeMsg {
	return &ChannelProposeMsg{
		ChannelID:           host.escrowKeyPair.Address(),
		GuestEscrowPubKey:   guestEscrowPubKey,
		HostRatchetAccount:  host.hostRatchetAccount.keyPair.Address(),
		GuestRatchetAccount: host.guestRatchetAccount.keyPair.Address(),
		MaxRoundDuration:    defaultMaxRoundDuration,
		FinalityDelay:       defaultFinalityDelay,
		Feerate:             defaultFeerate,
		HostAmount:          defaultHostAmount,
		FundingTime:         getBlockChainTime(),
		HostAccount:         host.selfKeyPair.Address(),
	}
}

func (host *hostAccount) createPaymentProposeMsg(roundNumber int, guestAddress string) (*PaymentProposeMsg, error) {
	rsn := roundSequenceNumber(host.baseSequenceNumber, roundNumber)
	paymentTime := getBlockChainTime()

	settleWithGuestTx, err := host.createAndSignSettleWithGuestTx(uint64(rsn), paymentTime, defaultPaymentAmount)
	if err != nil {
		return nil, err
	}

	settleWithHostTx, err := host.createAndSignSettleWithHostTx(uint64(rsn), paymentTime)
	if err != nil {
		return nil, err
	}

	msg := &PaymentProposeMsg{
		ChannelID:                host.escrowKeyPair.Address(),
		RoundNumber:              roundNumber,
		PaymentTime:              paymentTime,
		PaymentAmount:            defaultPaymentAmount,
		SenderSettleWithGuestSig: &settleWithGuestTx.E.Signatures[0],
		SenderSettleWithHostSig:  &settleWithHostTx.E.Signatures[0],
	}
	host.cache.paymentProposeMsg = msg
	return msg, nil
}

func (host *hostAccount) loadSequenceNumber() int {
	sequenceNumber, err := loadSequenceNumber(host.selfKeyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}

func (host *hostAccount) loadBalance() string {
	return loadBalance(host.selfKeyPair.Address())
}
