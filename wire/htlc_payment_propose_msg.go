package wire

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type HTLCPaymentProposeMsg struct {
	ChannelID             string
	RoundNumber           int
	PaymentTime           uint64
	PaymentAmount         string
	RHash                 [sha256.Size]byte
	HtlcResolutionAddress string
	// SenderSettleWithGuestSig *xdr.DecoratedSignature // (or empty)
	// SenderSettleWithHostSig  *xdr.DecoratedSignature
}

func (msg *HTLCPaymentProposeMsg) String() string {
	tmpl := `
################### HTLCPaymentProposeMsg |Host -> Guest| ###############
	ChannelID                %v
	RoundNumber              %v
	PaymentTime              %v
	PaymentAmount            %v
	RHash                    %v
	HtlcResolutionAddress    %v
#####################################################################
	`
	return fmt.Sprintf(
		tmpl,
		msg.ChannelID,
		msg.RoundNumber,
		msg.PaymentTime,
		msg.PaymentAmount,
		hex.EncodeToString(msg.RHash[:]),
		msg.HtlcResolutionAddress,
	)
}
