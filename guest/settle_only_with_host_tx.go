package guest

import (
	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
)

func (guest *Account) createAndSignSettleOnlyWithHostTx(
	fundingTime uint64,
	roundNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := script_utils.CreateSettleOnlyWithHostTx(
		guest.cache.channelProposeMsg.HostAccount,
		guest.cache.channelProposeMsg.ChannelID,
		guest.cache.channelProposeMsg.GuestRatchetAccount,
		guest.cache.channelProposeMsg.HostRatchetAccount,
		fundingTime,
		tools.RoundSequenceNumber(guest.baseSequenceNumber, roundNumber),
	)

	txe, err := tx.Sign(guest.KeyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}
