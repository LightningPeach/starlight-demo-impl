package wire

import (
	"fmt"

	"github.com/stellar/go/xdr"
)

type HTLCPaymentAcceptMsg struct {
	ChannelID                                   string
	RoundNumber                                 int
	RecipientRatchetSig                         *xdr.DecoratedSignature
	RecipientSettleOnlyWithHostAndActiveHtlcSig *xdr.DecoratedSignature
	RecipientHtlcTimeoutSig                     *xdr.DecoratedSignature
}

func (msg *HTLCPaymentAcceptMsg) String() string {
	tmpl := `
############### HTLCPaymentAcceptMsg |Guest -> Host| ###############
	ChannelID                                   %v
	RoundNumber                                 %v
	RecipientRatchetSig                         %v
	RecipientSettleOnlyWithHostAndActiveHtlcSig %v
	RecipientHtlcTimeoutSig                     %v
####################################################################
	`
	return fmt.Sprintf(
		tmpl,
		msg.ChannelID,
		msg.RoundNumber,
		msg.RecipientRatchetSig,
		msg.RecipientSettleOnlyWithHostAndActiveHtlcSig,
		msg.RecipientHtlcTimeoutSig,
	)
}
