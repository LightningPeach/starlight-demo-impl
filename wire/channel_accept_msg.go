package wire

import (
	"fmt"

	"github.com/stellar/go/xdr"
)

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