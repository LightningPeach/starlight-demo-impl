package wire

import (
	"fmt"
	"github.com/stellar/go/xdr"
)

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