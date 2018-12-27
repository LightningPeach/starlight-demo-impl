package main

import (
	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

type ChannelProposeMsg struct {
	ChannelID           string // the same as HostEscrowPubKey and the account ID of EscrowAccount
	GuestEscrowPubKey   string
	HostRatchetAccount  string
	GuestRatchetAccount string
	MaxRoundDuration    uint64
	FinalityDelay       uint64
	Feerate             string // TODO(evg): what is it?
	HostAmount          string
	FundingTime         uint64
	HostAccount         string
}

type ChannelAcceptMsg struct {
	ChannelID                  string
	GuestRatchetRound1Sig      *xdr.DecoratedSignature // TODO(evg): use only sig instead of whole signed tx
	GuestSettleOnlyWithHostSig *xdr.DecoratedSignature // TODO(evg): use only sig instead of whole signed tx
}

type PaymentProposeMsg struct {
	ChannelID                string
	RoundNumber              int
	PaymentTime              uint64
	PaymentAmount            string
	SenderSettleWithGuestSig *build.TransactionEnvelopeBuilder // (or empty)
	SenderSettleWithHostSig  *build.TransactionEnvelopeBuilder
}

type PaymentAcceptMsg struct {
	ChannelID                   string
	RoundNumber                 int
	RecipientRatchetSig         *build.TransactionEnvelopeBuilder
	RecipientSettleWithGuestSig *build.TransactionEnvelopeBuilder
	RecipientSettleWithHostSig  *build.TransactionEnvelopeBuilder
}
