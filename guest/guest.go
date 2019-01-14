package guest

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/LightningPeach/starlight-demo-impl/wire"
	"github.com/stellar/go/keypair"
)

type guestMessageCache struct {
	channelProposeMsg *wire.ChannelProposeMsg
}

type Account struct {
	KeyPair *keypair.Full

	baseSequenceNumber int

	cache *guestMessageCache
}

func New() (*Account, error) {
	fmt.Println("creating guest account:")
	keyPair, err := tools.CreateAccount()
	if err != nil {
		return nil, err
	}
	guestAccount := &Account{
		KeyPair: keyPair,
		cache:   new(guestMessageCache),
	}
	fmt.Printf("balance: %v\n\n", tools.LoadBalance(guestAccount.KeyPair.Address()))
	return guestAccount, nil
}

func (guest *Account) AddInvoice() [sha256.Size]byte {
	debugR := [32]byte{}
	rHash := sha256.Sum256(debugR[:])
	return rHash
}

func (guest *Account) loadSequenceNumber() int {
	sequenceNumber, err := tools.LoadSequenceNumber(guest.KeyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}
