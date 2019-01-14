package guest

import (
	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

func (guest *Account) CreateAndSignHtlcSuccessTx(htlcResolutionAddress string, paymentTime uint64) (*build.TransactionEnvelopeBuilder, error) {
	tx, err := script_utils.CreateHtlcSuccessTx(htlcResolutionAddress, guest.KeyPair.Address())
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(guest.KeyPair.Seed())
	if err != nil {
		return nil, err
	}

	hint := [4]byte{}
	copy(hint[:], script_utils.RHash[28:])

	txe.E.Signatures = append(txe.E.Signatures, xdr.DecoratedSignature{
		Hint:      hint,
		Signature: script_utils.RPreImage,
	})

	return &txe, nil
}
