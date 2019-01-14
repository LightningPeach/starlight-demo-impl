package guest

import (
	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
)

func (guest *Account) createAndSignRatchetTxForHost(
	paymentTime uint64, // payment of funding time
	roundNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {
	tx, err := script_utils.CreateRatchetTx(
		guest.cache.channelProposeMsg.HostRatchetAccount,
		guest.cache.channelProposeMsg.ChannelID,
		paymentTime,
		tools.RoundSequenceNumber(guest.baseSequenceNumber, roundNumber),
	)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(guest.KeyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}
