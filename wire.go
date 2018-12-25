package main

import "github.com/stellar/go/build"

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
	GuestRatchetRound1Sig      *build.TransactionEnvelopeBuilder // TODO(evg): use only sig instead of whole signed tx
	GuestSettleOnlyWithHostSig string
}