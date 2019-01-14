package host

import (
	"crypto/sha256"

	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/LightningPeach/starlight-demo-impl/wire"
)

func (host *Account) CreateChannelProposeMsg(guestEscrowPubKey string) *wire.ChannelProposeMsg {
	return &wire.ChannelProposeMsg{
		ChannelID:             host.escrowKeyPair.Address(),
		GuestEscrowPubKey:     guestEscrowPubKey,
		HostRatchetAccount:    host.hostRatchetAccount.KeyPair.Address(),
		GuestRatchetAccount:   host.guestRatchetAccount.KeyPair.Address(),
		MaxRoundDuration:      tools.DefaultMaxRoundDuration,
		FinalityDelay:         tools.DefaultFinalityDelay,
		Feerate:               tools.DefaultFeerate,
		HostAmount:            tools.DefaultHostAmount,
		FundingTime:           tools.GetBlockChainTime(),
		HostAccount:           host.selfKeyPair.Address(),
		HTLCResolutionAccount: host.HtlcResolutionAccount.KeyPair.Address(),
	}
}

func (host *Account) CreatePaymentProposeMsg(roundNumber int, guestAddress string) (*wire.PaymentProposeMsg, error) {
	rsn := tools.RoundSequenceNumber(host.BaseSequenceNumber, roundNumber)
	paymentTime := tools.GetBlockChainTime()

	settleWithGuestTx, err := host.CreateAndSignSettleWithGuestTx(uint64(rsn), paymentTime, tools.DefaultPaymentAmount)
	if err != nil {
		return nil, err
	}

	settleWithHostTx, err := host.createAndSignSettleWithHostTx(uint64(rsn), paymentTime)
	if err != nil {
		return nil, err
	}

	msg := &wire.PaymentProposeMsg{
		ChannelID:                host.escrowKeyPair.Address(),
		RoundNumber:              roundNumber,
		PaymentTime:              paymentTime,
		PaymentAmount:            tools.DefaultPaymentAmount,
		SenderSettleWithGuestSig: &settleWithGuestTx.E.Signatures[0],
		SenderSettleWithHostSig:  &settleWithHostTx.E.Signatures[0],
	}
	host.cache.paymentProposeMsg = msg
	return msg, nil
}

func (host *Account) CreateHTLCPaymentProposeMsg(
	roundNumber int,
	guestAddress string,
	rHash [sha256.Size]byte,
) (*wire.HTLCPaymentProposeMsg, error) {

	paymentTime := tools.GetBlockChainTime()

	msg := &wire.HTLCPaymentProposeMsg{
		ChannelID:             host.escrowKeyPair.Address(),
		RoundNumber:           roundNumber,
		PaymentTime:           paymentTime,
		PaymentAmount:         tools.DefaultPaymentAmount,
		RHash:                 rHash,
		HtlcResolutionAddress: host.HtlcResolutionAccount.KeyPair.Address(),
	}
	host.cache.htlcPaymentProposeMsg = msg
	return msg, nil
}
