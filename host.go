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

type hostAccount struct {
	selfKeyPair *keypair.Full

	escrowKeyPair       *keypair.Full
	hostRatchetAccount  *ratchetAccount
	guestRatchetAccount *ratchetAccount

	baseSequenceNumber int
}

func newHostAccount() (*hostAccount, error) {
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

	return &hostAccount{
		selfKeyPair:         selfKeyPair,
		escrowKeyPair:       escrowKeyPair,
		hostRatchetAccount:  hostRatchetAccount,
		guestRatchetAccount: guestRatchetAccount,
	}, nil
}

func (host *hostAccount) setupAccountTx(account accountType) error {
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
		// build.AutoSequence{horizon.DefaultTestNetClient},
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
	return nil
}

// TODO(evg): minTime/maxTime
func (host *hostAccount) fundingTx(guestEscrowPubKey string) error {
	// TODO(evg): use blockchain timestamp instead of system time
	fundingTime := time.Now().Unix()
	// TODO(evg): adjust constants
	const (
		maxRoundDuration = 3600
		finalityDelay    = 3600
	)

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.selfKeyPair.Address()},
		build.Sequence{Sequence: uint64(host.loadSequenceNumber()) + 1},
		build.Timebounds{
			MaxTime: uint64(fundingTime + maxRoundDuration + finalityDelay),
		},
		// Escrow Account
		build.Payment(
			build.Destination{AddressOrSeed: host.escrowKeyPair.Address()},
			// build.NativeAmount{Amount: strconv.Itoa(0.5 + 8 * feeRate)},
			build.NativeAmount{Amount: "500.50008"}, // defaultHostAmount + 0.5 + 8 * feerate // TODO(evg): refactor it
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
			build.NativeAmount{Amount: "1.00001"},
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
			build.NativeAmount{Amount: "0.50001"},
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

//func (host *hostAccount) ratchetTx(ratchetTx *build.TransactionEnvelopeBuilder) error {
//	if err := ratchetTx.Mutate(build.Sign{Seed: host.escrowKeyPair.Seed()}); err != nil {
//		return err
//	}
//	//if err := ratchetTx.Mutate(build.Sign{Seed: host.hostRatchetAccount.keyPair.Seed()}); err != nil {
//	//	return err
//	//}
//
//	if err := host.publishTx(ratchetTx); err != nil {
//		fmt.Println("tx fail")
//		err2 := err.(*horizon.Error).Problem
//		fmt.Println("Type: ", err2.Type)
//		fmt.Println("Title: ", err2.Title)
//		fmt.Println("Status: ", err2.Status)
//		fmt.Println("Detail:", err2.Detail)
//		fmt.Println("Instance: ", err2.Instance)
//		for key, value := range err2.Extras {
//			fmt.Println("KEYVALUE: ", key, string(value))
//		}
//		// fmt.Println("Extras: ",   err2.Extras)
//		return err
//	}
//	return nil
//}
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

func (host *hostAccount) settleOnlyWithHostTx(settleOnlyWithHostTx *build.TransactionEnvelopeBuilder) error {
	if err := settleOnlyWithHostTx.Mutate(build.Sign{Seed: host.escrowKeyPair.Seed()}); err != nil {
		return err
	}
	// TODO(evg): remove it signer??
	//if err := settleOnlyWithHostTx.Mutate(build.Sign{Seed: host.hostRatchetAccount.keyPair.Seed()}); err != nil {
	//	return err
	//}
	//if err := settleOnlyWithHostTx.Mutate(build.Sign{Seed: host.guestRatchetAccount.keyPair.Seed()}); err != nil {
	//	return err
	//}

	if err := host.publishTx(settleOnlyWithHostTx); err != nil {
		fmt.Println("tx fail")
		err2 := err.(*horizon.Error).Problem
		fmt.Println("Type: ", err2.Type)
		fmt.Println("Title: ", err2.Title)
		fmt.Println("Status: ", err2.Status)
		fmt.Println("Detail:", err2.Detail)
		fmt.Println("Instance: ", err2.Instance)
		for key, value := range err2.Extras {
			fmt.Println("KEYVALUE: ", key, string(value))
		}
		// fmt.Println("Extras: ",   err2.Extras)
		return err
	}
	return nil
}

// TODO(evg): try to submit it
func (host *hostAccount) createSettleWithGuestTx(
	rsn,
	paymentTime uint64,
	guestAddress,
	guestAmount string,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.escrowKeyPair.Address()},
		build.Sequence{Sequence: rsn + 2},
		build.Timebounds{
			MinTime: paymentTime + 2*defaultFinalityDelay + defaultMaxRoundDuration,
		},
		// Pay GuestAmount from EscrowAccount to GuestAccount
		build.Payment(
			build.Destination{AddressOrSeed: guestAddress},
			build.NativeAmount{Amount: guestAmount},
		),
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

// TODO(evg): try to submit it
func (host *hostAccount) createSettleWithHostTx(rsn, paymentTime uint64) (*build.TransactionEnvelopeBuilder, error) {
	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.escrowKeyPair.Address()},
		build.Sequence{Sequence: rsn + 3},
		build.Timebounds{
			MinTime: paymentTime + 2*defaultFinalityDelay + defaultMaxRoundDuration,
		},
		//Merge account EscrowAccount to HostAccount
		//Merge account GuestRatchetAccount to HostAccount
		//Merge account HostRatchetAccount to HostAccount
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: host.escrowKeyPair.Address()},
			build.Destination{AddressOrSeed: host.selfKeyPair.Address()},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: host.guestRatchetAccount.keyPair.Address()},
			build.Destination{AddressOrSeed: host.selfKeyPair.Address()},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: host.hostRatchetAccount.keyPair.Address()},
			build.Destination{AddressOrSeed: host.selfKeyPair.Address()},
		),
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
	resp, err := horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		return err
	}

	//fmt.Println("Successful Transaction:")
	//fmt.Println("Ledger:", resp.Ledger)
	//fmt.Println("Hash:", resp.Hash)
	_ = resp

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
	//  rsn,
	//	paymentTime uint64,
	//	guestAddress,
	//	guestAmount string,
	rsn := roundSequenceNumber(host.baseSequenceNumber, roundNumber)
	paymentTime := getBlockChainTime()

	settleWithGuestTx, err := host.createSettleWithGuestTx(uint64(rsn), paymentTime, guestAddress, defaultPaymentAmount)
	if err != nil {
		return nil, err
	}

	// rsn,
	// paymentTime uint64
	settleWithHostTx, err := host.createSettleWithHostTx(uint64(rsn), paymentTime)
	if err != nil {
		return nil, err
	}

	return &PaymentProposeMsg{
		ChannelID:                host.selfKeyPair.Address(),
		RoundNumber:              roundNumber,
		PaymentTime:              paymentTime,
		PaymentAmount:            defaultPaymentAmount,
		SenderSettleWithGuestSig: settleWithGuestTx,
		SenderSettleWithHostSig:  settleWithHostTx,
	}, nil
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
