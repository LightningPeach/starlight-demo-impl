package guest

import (
	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/LightningPeach/starlight-demo-impl/wire"
)

func (guest *Account) ReceiveChannelProposeMsg(msg *wire.ChannelProposeMsg) (*wire.ChannelAcceptMsg, error) {
	guest.cache.channelProposeMsg = &*msg

	baseSequenceNumber, err := tools.LoadSequenceNumber(msg.ChannelID)
	if err != nil {
		return nil, err
	}
	guest.baseSequenceNumber = baseSequenceNumber

	ratchetTx, err := guest.createAndSignRatchetTxForHost(msg.FundingTime, 1)
	if err != nil {
		return nil, err
	}

	settleOnlyWithHostTx, err := guest.createAndSignSettleOnlyWithHostTx(msg.FundingTime, 1)

	return &wire.ChannelAcceptMsg{
		ChannelID:                  msg.ChannelID,
		GuestRatchetRound1Sig:      &ratchetTx.E.Signatures[0],
		GuestSettleOnlyWithHostSig: &settleOnlyWithHostTx.E.Signatures[0],
	}, nil
}

func (guest *Account) ReceivePaymentProposeMsg(msg *wire.PaymentProposeMsg) (*wire.PaymentAcceptMsg, error) {
	rsn := tools.RoundSequenceNumber(int(guest.baseSequenceNumber), msg.RoundNumber)
	ratchetTxForOffChainPayment, err := guest.createAndSignRatchetTxForHost(msg.PaymentTime, msg.RoundNumber)
	if err != nil {
		return nil, err
	}

	tx, err := script_utils.CreateSettleWithGuestTx(
		uint64(rsn),
		msg.PaymentTime,
		guest.KeyPair.Address(),
		msg.PaymentAmount,
		guest.cache.channelProposeMsg.ChannelID,
	)
	if err != nil {
		return nil, err
	}

	txeGuest, err := tx.Sign(guest.KeyPair.Seed())
	if err != nil {
		return nil, err
	}

	tx, err = script_utils.CreateSettleWithHostTx(
		uint64(rsn),
		msg.PaymentTime,
		msg.ChannelID,
		guest.cache.channelProposeMsg.GuestRatchetAccount,
		guest.cache.channelProposeMsg.HostRatchetAccount,
		guest.cache.channelProposeMsg.HostAccount,
		guest.cache.channelProposeMsg.HTLCResolutionAccount,
	)
	if err != nil {
		return nil, err
	}

	txeHost, err := tx.Sign(guest.KeyPair.Seed())
	if err != nil {
		return nil, err
	}
	txeHost.E.Signatures = append(txeHost.E.Signatures, *msg.SenderSettleWithHostSig)

	return &wire.PaymentAcceptMsg{
		ChannelID:                   msg.ChannelID,
		RoundNumber:                 msg.RoundNumber,
		RecipientRatchetSig:         &ratchetTxForOffChainPayment.E.Signatures[0],
		RecipientSettleWithGuestSig: &txeGuest.E.Signatures[0],
		RecipientSettleWithHostSig:  &txeHost.E.Signatures[0],
	}, nil
}

func (guest *Account) ReceiveHTLCPaymentProposeMsg(msg *wire.HTLCPaymentProposeMsg) (*wire.HTLCPaymentAcceptMsg, error) {
	rsn := tools.RoundSequenceNumber(guest.baseSequenceNumber, msg.RoundNumber)
	ratchetTxForOffChainPayment, err := guest.createAndSignRatchetTxForHost(msg.PaymentTime, msg.RoundNumber)
	if err != nil {
		return nil, err
	}

	settleOnlyWithHostAndActiveHtlcTx, err := script_utils.CreateSettleOnlyWithHostAndActiveHtlcTx(
		guest.cache.channelProposeMsg.HostAccount,
		msg.ChannelID,
		guest.cache.channelProposeMsg.GuestRatchetAccount,
		guest.cache.channelProposeMsg.HostRatchetAccount,
		msg.PaymentTime,
		rsn,
		guest.cache.channelProposeMsg.HTLCResolutionAccount,
		msg.PaymentAmount,
	)
	if err != nil {
		return nil, err
	}
	txe, err := settleOnlyWithHostAndActiveHtlcTx.Sign(guest.KeyPair.Seed())
	if err != nil {
		return nil, err
	}

	htlcTimeoutTx, err := script_utils.CreateHtlcTimeoutTx(msg.HtlcResolutionAddress, guest.cache.channelProposeMsg.HostAccount, msg.PaymentTime)
	if err != nil {
		return nil, err
	}
	htlcTimeoutTxe, err := htlcTimeoutTx.Sign(guest.KeyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &wire.HTLCPaymentAcceptMsg{
		ChannelID:           msg.ChannelID,
		RoundNumber:         msg.RoundNumber,
		RecipientRatchetSig: &ratchetTxForOffChainPayment.E.Signatures[0],
		RecipientSettleOnlyWithHostAndActiveHtlcSig: &txe.E.Signatures[0],
		RecipientHtlcTimeoutSig:                     &htlcTimeoutTxe.E.Signatures[0],
	}, nil
}
