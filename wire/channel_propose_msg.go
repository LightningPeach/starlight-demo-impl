package wire

import "fmt"

type ChannelProposeMsg struct {
	ChannelID             string // the same as HostEscrowPubKey and the account ID of EscrowAccount
	GuestEscrowPubKey     string
	HostRatchetAccount    string
	GuestRatchetAccount   string
	MaxRoundDuration      uint64
	FinalityDelay         uint64
	Feerate               string
	HostAmount            string
	FundingTime           uint64
	HostAccount           string
	HTLCResolutionAccount string
}

func (msg *ChannelProposeMsg) String() string {
	tmpl := `
################### ChannelProposeMsg |Host -> Guest| ###############
	ChannelID             %v
	GuestEscrowPubKey     %v
	HostRatchetAccount    %v
	GuestRatchetAccount   %v
	MaxRoundDuration      %v
	FinalityDelay         %v
	Feerate               %v
	HostAmount            %v
	FundingTime           %v
	HostAccount           %v
	HTLCResolutionAccount %v
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
		msg.HTLCResolutionAccount,
	)
}