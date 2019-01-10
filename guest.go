package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/stellar/go/build"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"
	"log"
)

type guestMessageCache struct {
	channelProposeMsg *ChannelProposeMsg
}

type guestAccount struct {
	keyPair *keypair.Full

	baseSequenceNumber int

	cache *guestMessageCache
}

func newGuestAccount() (*guestAccount, error) {
	fmt.Println("creating guest account:")
	keyPair, err := createAccount()
	if err != nil {
		return nil, err
	}
	guestAccount := &guestAccount{
		keyPair: keyPair,
		cache:   new(guestMessageCache),
	}
	fmt.Printf("balance: %v\n\n", loadBalance(guestAccount.keyPair.Address()))
	return guestAccount, nil
}

func (guest *guestAccount) createAndSignRatchetTxForHost(
	paymentTime uint64, // payment of funding time
	roundNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {
	tx, err := createRatchetTx(
		guest.cache.channelProposeMsg.HostRatchetAccount,
		guest.cache.channelProposeMsg.ChannelID,
		paymentTime,
		roundSequenceNumber(guest.baseSequenceNumber, roundNumber),
	)
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
	fundingTime uint64,
	roundNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := createSettleOnlyWithHostTx(
		guest.cache.channelProposeMsg.HostAccount,
		guest.cache.channelProposeMsg.ChannelID,
		guest.cache.channelProposeMsg.GuestRatchetAccount,
		guest.cache.channelProposeMsg.HostRatchetAccount,
		fundingTime,
		roundSequenceNumber(guest.baseSequenceNumber, roundNumber),
		guest.cache.channelProposeMsg.HTLCResolutionAccount,
	)

	txe, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}

func (guest *guestAccount) createAndSignHtlcSuccessTx(htlcResolutionAddress string, paymentTime uint64) (*build.TransactionEnvelopeBuilder, error) {
	tx, err := createHtlcTimeoutTx(htlcResolutionAddress, guest.keyPair.Address(), paymentTime)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}

	hint := [4]byte{}
	copy(hint[:], rHash[28:])

	txe.E.Signatures = append(txe.E.Signatures, xdr.DecoratedSignature{
		Hint: hint,
		Signature: rPreImage,
	})

	return &txe, nil
}

func (guest *guestAccount) receiveChannelProposeMsg(msg *ChannelProposeMsg) (*ChannelAcceptMsg, error) {
	guest.cache.channelProposeMsg = &*msg

	baseSequenceNumber, err := loadSequenceNumber(msg.ChannelID)
	if err != nil {
		return nil, err
	}
	guest.baseSequenceNumber = baseSequenceNumber
	// rsn := roundSequenceNumber(baseSequenceNumber, 1)

	ratchetTx, err := guest.createAndSignRatchetTxForHost(msg.FundingTime, 1)
	if err != nil {
		return nil, err
	}

	settleOnlyWithHostTx, err := guest.createAndSignSettleOnlyWithHostTx(msg.FundingTime, 1)

	return &ChannelAcceptMsg{
		ChannelID:                  msg.ChannelID,
		GuestRatchetRound1Sig:      &ratchetTx.E.Signatures[0],
		GuestSettleOnlyWithHostSig: &settleOnlyWithHostTx.E.Signatures[0],
	}, nil
}

func (guest *guestAccount) receivePaymentProposeMsg(msg *PaymentProposeMsg) (*PaymentAcceptMsg, error) {
	rsn := roundSequenceNumber(int(guest.baseSequenceNumber), msg.RoundNumber)
	ratchetTxForOffChainPayment, err := guest.createAndSignRatchetTxForHost(msg.PaymentTime, msg.RoundNumber)
	if err != nil {
		return nil, err
	}

	tx, err := createSettleWithGuestTx(
		uint64(rsn),
		msg.PaymentTime,
		guest.keyPair.Address(),
		msg.PaymentAmount,
		guest.cache.channelProposeMsg.ChannelID,
	)
	if err != nil {
		return nil, err
	}

	txeGuest, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}
	// txeGuest.E.Signatures = append(txeGuest.E.Signatures, *msg.SenderSettleWithGuestSig)

	//rsn,
	//paymentTime uint64,
	//escrowAddress,
	//guestRatchetAddress,
	//hostRatchetAddress,
	//hostAddress string,
	tx, err = createSettleWithHostTx(
		uint64(rsn),
		msg.PaymentTime,
		msg.ChannelID,
		guest.cache.channelProposeMsg.GuestRatchetAccount,
		guest.cache.channelProposeMsg.HostRatchetAccount,
		guest.cache.channelProposeMsg.HostAccount,
		guest.cache.channelProposeMsg.HTLCResolutionAccount,
	)
	if err != nil {
		return nil, err
	}

	txeHost, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}
	txeHost.E.Signatures = append(txeHost.E.Signatures, *msg.SenderSettleWithHostSig)

	//copyTxHost := msg.SenderSettleWithHostSig
	//copyTxHost.Mutate(build.Sign{Seed: guest.keyPair.Seed()})

	return &PaymentAcceptMsg{
		ChannelID:                   msg.ChannelID,
		RoundNumber:                 msg.RoundNumber,
		RecipientRatchetSig:         &ratchetTxForOffChainPayment.E.Signatures[0],
		RecipientSettleWithGuestSig: &txeGuest.E.Signatures[0],
		RecipientSettleWithHostSig:  &txeHost.E.Signatures[0],
	}, nil
}

func (guest *guestAccount) receiveHTLCPaymentProposeMsg(msg *HTLCPaymentProposeMsg) (*HTLCPaymentAcceptMsg, error) {
	rsn := roundSequenceNumber(guest.baseSequenceNumber, msg.RoundNumber)
	ratchetTxForOffChainPayment, err := guest.createAndSignRatchetTxForHost(msg.PaymentTime, msg.RoundNumber)
	if err != nil {
		return nil, err
	}

	settleOnlyWithHostAndActiveHtlcTx, err := createSettleOnlyWithHostAndActiveHtlcTx(
		guest.cache.channelProposeMsg.HostAccount,
		msg.ChannelID,
		guest.cache.channelProposeMsg.GuestRatchetAccount,
		guest.cache.channelProposeMsg.HostRatchetAccount,
		msg.PaymentTime,
		rsn,
		guest.cache.channelProposeMsg.HTLCResolutionAccount,
		msg.PaymentAmount,
	)
	if err != nil {
		return nil, err
	}
	txe, err := settleOnlyWithHostAndActiveHtlcTx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}

	htlcTimeoutTx, err := createHtlcTimeoutTx(msg.HtlcResolutionAddress, guest.cache.channelProposeMsg.HostAccount, msg.PaymentTime)
	if err != nil {
		return nil, err
	}
	htlcTimeoutTxe, err := htlcTimeoutTx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &HTLCPaymentAcceptMsg{
		ChannelID:           msg.ChannelID,
		RoundNumber:         msg.RoundNumber,
		RecipientRatchetSig: &ratchetTxForOffChainPayment.E.Signatures[0],
		RecipientSettleOnlyWithHostAndActiveHtlcSig: &txe.E.Signatures[0],
		RecipientHtlcTimeoutSig:                     &htlcTimeoutTxe.E.Signatures[0],
	}, nil
}

func (guest *guestAccount) addInvoice() [sha256.Size]byte {
	debugR := [32]byte{}
	rHash := sha256.Sum256(debugR[:])
	return rHash
}

func (guest *guestAccount) loadSequenceNumber() int {
	sequenceNumber, err := loadSequenceNumber(guest.keyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}
