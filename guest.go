package main

import (
	"fmt"
	"github.com/stellar/go/build"
	"github.com/stellar/go/keypair"
	"log"
)

type guestAccount struct {
	keyPair *keypair.Full
}

func newGuestAccount() (*guestAccount, error) {
	fmt.Println("creating guest account:")
	keyPair, err := createAccount()
	if err != nil {
		return nil, err
	}
	guestAccount := &guestAccount{
		keyPair: keyPair,
	}
	fmt.Printf("balance: %v\n\n", loadBalance(guestAccount.keyPair.Address()))
	return guestAccount, nil
}

func (guest *guestAccount) createAndSignRatchetTxForHost(
	escrowAddress,
	hostRatchetAddress string,
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {
	tx, err := createRatchetTx(hostRatchetAddress, escrowAddress, paymentTime, roundSequenceNumber)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}

func (guest *guestAccount) createAndSignSettleOnlyWithHostTx(
	hostAddress,
	escrowAddress,
	guestRatchetAddress,
	hostRatchetAddress string,
	fundingTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := createSettleOnlyWithHostTx(
		hostAddress,
		escrowAddress,
		guestRatchetAddress,
		hostRatchetAddress,
		fundingTime,
		roundSequenceNumber,
	)
	fmt.Println("createSettleOnlyWithHostTx.MinTime", fundingTime+2*defaultFinalityDelay+defaultMaxRoundDuration)

	txe, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}

func (guest *guestAccount) receiveChannelProposeMsg(msg *ChannelProposeMsg) (*ChannelAcceptMsg, error) {
	baseSequenceNumber, err := loadSequenceNumber(msg.ChannelID)
	if err != nil {
		return nil, err
	}
	rsn := roundSequenceNumber(baseSequenceNumber, 1)

	fmt.Println("createRatchetTx for host")
	ratchetTx, err := guest.createAndSignRatchetTxForHost(msg.HostRatchetAccount, msg.ChannelID, msg.FundingTime, rsn)
	if err != nil {
		return nil, err
	}

	fmt.Println("createSettleOnlyWithHostTx")
	settleOnlyWithHostTx, err := guest.createAndSignSettleOnlyWithHostTx(
		msg.HostAccount,
		msg.ChannelID,
		msg.GuestRatchetAccount,
		msg.HostRatchetAccount,
		msg.FundingTime,
		rsn,
	)

	return &ChannelAcceptMsg{
		ChannelID:                  msg.ChannelID,
		GuestRatchetRound1Sig:      &ratchetTx.E.Signatures[0],
		GuestSettleOnlyWithHostSig: &settleOnlyWithHostTx.E.Signatures[0],
	}, nil
}

//func (guest *guestAccount) createRatchetTxForOffChainPayment(
//	escrowAddress,
//	hostRatchetAddress string,
//	paymentTime uint64,
//	rsn int64,
//) (
//	*build.TransactionEnvelopeBuilder,
//	error,
//) {
//
//	sequence, err := loadSequenceNumber(hostRatchetAddress)
//	if err != nil {
//		return nil, err
//	}
//
//	tx, err := build.Transaction(
//		build.TestNetwork,
//		build.SourceAccount{AddressOrSeed: hostRatchetAddress},
//		// RatchetAccount.SequenceNumber + 1
//		build.Sequence{Sequence: uint64(sequence) + 1},
//		build.Timebounds{
//			MaxTime: paymentTime + defaultFinalityDelay + defaultMaxRoundDuration,
//		},
//		// Bump sequence of EscrowAccount to RoundSequenceNumber + 1
//		build.BumpSequence(
//			build.SourceAccount{AddressOrSeed: escrowAddress},
//			build.BumpTo(rsn+1),
//		),
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	txe, err := tx.Sign(guest.keyPair.Seed())
//	if err != nil {
//		return nil, err
//	}
//
//	return &txe, nil
//}

//rsn,
//paymentTime uint64,
//guestAddress,
//guestAmount,
//escrowAddress string,
//func (guest *guestAccount) createSettleWithGuestTx() {
//	createSettleWithGuestTx()
//}

// TODO(evg): guest should already know all args
func (guest *guestAccount) receivePaymentProposeMsg(
	msg *PaymentProposeMsg,
	escrowAddress,
	hostRatchetAddress string,
	bsn int64,
) (*PaymentAcceptMsg, error) {
	// escrowAddress,
	// hostRatchetAddress string,
	// paymentTime uint64,
	// rsn int64,
	rsn := roundSequenceNumber(int(bsn), msg.RoundNumber)
	ratchetTxForOffChainPayment, err := guest.createAndSignRatchetTxForHost(escrowAddress, hostRatchetAddress, msg.PaymentTime, rsn)
	if err != nil {
		return nil, err
	}

	//rsn,
	//paymentTime uint64,
	//guestAddress,
	//guestAmount,
	//escrowAddress string,
	tx, err := createSettleWithGuestTx(
		uint64(rsn),
		msg.PaymentTime,
		guest.keyPair.Address(),
		msg.PaymentAmount,
		escrowAddress,
	)

	txeGuest, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}
	txeGuest.E.Signatures = append(txeGuest.E.Signatures, *msg.SenderSettleWithGuestSig)
	//copyTxGuest := msg.SenderSettleWithGuestSig
	//copyTxGuest.Mutate(build.Sign{Seed: guest.keyPair.Seed()})

	copyTxHost := msg.SenderSettleWithHostSig
	copyTxHost.Mutate(build.Sign{Seed: guest.keyPair.Seed()})

	return &PaymentAcceptMsg{
		ChannelID:                   msg.ChannelID,
		RoundNumber:                 msg.RoundNumber,
		RecipientRatchetSig:         ratchetTxForOffChainPayment,
		RecipientSettleWithGuestSig: &txeGuest,
		RecipientSettleWithHostSig:  copyTxHost,
	}, nil
}

func (guest *guestAccount) loadSequenceNumber() int {
	sequenceNumber, err := loadSequenceNumber(guest.keyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}
