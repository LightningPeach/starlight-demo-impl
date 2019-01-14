package wire

import (
	"fmt"

	"github.com/stellar/go/xdr"
)

type PaymentAcceptMsg struct {
	ChannelID                   string
	RoundNumber                 int
	RecipientRatchetSig         *xdr.DecoratedSignature
	RecipientSettleWithGuestSig *xdr.DecoratedSignature
	RecipientSettleWithHostSig  *xdr.DecoratedSignature
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