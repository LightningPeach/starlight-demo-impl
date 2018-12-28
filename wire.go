package main

import (
	"fmt"
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

func (msg *ChannelProposeMsg) String() string {
	tmpl := `
################### ChannelProposeMsg |Host -> Guest| ###############
	ChannelID           %v
	GuestEscrowPubKey   %v
	HostRatchetAccount  %v
	GuestRatchetAccount %v
	MaxRoundDuration    %v
	FinalityDelay       %v
	Feerate             %v
	HostAmount          %v
	FundingTime         %v
	HostAccount         %v
#####################################################################
	`
	return fmt.Sprintf(
		tmpl,
		msg.ChannelID,
		msg.GuestEscrowPubKey,
		msg.HostRatchetAccount,
		msg.GuestRatchetAccount,
		msg.MaxRoundDuration,
		msg.FinalityDelay,
		msg.Feerate,
		msg.HostAmount,
		msg.FundingTime,
		msg.HostAccount,
	)
}

type ChannelAcceptMsg struct {
	ChannelID                  string
	GuestRatchetRound1Sig      *xdr.DecoratedSignature
	GuestSettleOnlyWithHostSig *xdr.DecoratedSignature
}

func (msg *ChannelAcceptMsg) String() string {
	tmpl := `
################### ChannelAcceptMsg |Guest -> Host| ###############
	ChannelID                  %v
	GuestRatchetRound1Sig      %v
	GuestSettleOnlyWithHostSig %v
####################################################################
	`
	return fmt.Sprintf(tmpl, msg.ChannelID, msg.GuestRatchetRound1Sig, msg.GuestSettleOnlyWithHostSig)
}

type PaymentProposeMsg struct {
	ChannelID                string
	RoundNumber              int
	PaymentTime              uint64
	PaymentAmount            string
	SenderSettleWithGuestSig *xdr.DecoratedSignature // (or empty)
	SenderSettleWithHostSig  *xdr.DecoratedSignature
}

func (msg *PaymentProposeMsg) String() string {
	tmpl := `
################### PaymentProposeMsg |Host -> Guest| ###############
	ChannelID                %v
	RoundNumber              %v
	PaymentTime              %v
	PaymentAmount            %v
	SenderSettleWithGuestSig %v
	SenderSettleWithHostSig  %v
#####################################################################
	`
	return fmt.Sprintf(
		tmpl,
		msg.ChannelID,
		msg.RoundNumber,
		msg.PaymentTime,
		msg.PaymentAmount,
		msg.SenderSettleWithGuestSig,
		msg.SenderSettleWithHostSig,
	)
}

type PaymentAcceptMsg struct {
	ChannelID                   string
	RoundNumber                 int
	RecipientRatchetSig         *xdr.DecoratedSignature
	RecipientSettleWithGuestSig *build.TransactionEnvelopeBuilder
	RecipientSettleWithHostSig  *build.TransactionEnvelopeBuilder
}

func (msg *PaymentAcceptMsg) String() string {
	tmpl := `
################### PaymentAcceptMsg |Guest -> Host| ###############
	ChannelID                   %v
	RoundNumber                 %v
	RecipientRatchetSig         %v
	RecipientSettleWithGuestSig %v
	RecipientSettleWithHostSig  %v
####################################################################
	`
	return fmt.Sprintf(
		tmpl,
		msg.ChannelID,
		msg.RoundNumber,
		msg.RecipientRatchetSig,
		msg.RecipientSettleWithGuestSig,
		msg.RecipientSettleWithHostSig,
	)
}