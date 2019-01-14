package demo

import (
	"fmt"
	"log"

	"github.com/LightningPeach/starlight-demo-impl/guest"
	"github.com/LightningPeach/starlight-demo-impl/host"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/LightningPeach/starlight-demo-impl/wire"
)

func Setup() (*host.Account, *guest.Account) {
	fmt.Println("starlight_demo")
	hostAccount, err := host.New()
	if err != nil {
		log.Fatal(err)
	}

	guestAccount, err := guest.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("creating channel accounts:")
	if err := hostAccount.PublishSetupAccountTx(tools.HostRatchetAccount); err != nil {
		log.Fatal(err)
	}
	if err := hostAccount.PublishSetupAccountTx(tools.GuestRatchetAccount); err != nil {
		log.Fatal(err)
	}
	if tools.HtlcMode {
		if err := hostAccount.PublishSetupAccountTx(tools.HtlcResolutionAccountType); err != nil {
			log.Fatal(err)
		}
	}
	if err := hostAccount.PublishSetupAccountTx(tools.EscrowAccount); err != nil {
		log.Fatal(err)
	}

	return hostAccount, guestAccount
}

func OpenChannel(hostAccount *host.Account, guestAccount *guest.Account) (*wire.ChannelProposeMsg, *wire.ChannelAcceptMsg) {
	channelProposeMsg := hostAccount.CreateChannelProposeMsg(guestAccount.KeyPair.Address())
	fmt.Println(channelProposeMsg)

	channelAcceptMsg, err := guestAccount.ReceiveChannelProposeMsg(channelProposeMsg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(channelAcceptMsg)

	if err := hostAccount.PublishFundingTx(guestAccount.KeyPair.Address()); err != nil {
		tools.ShowDetailError(err)
		log.Fatal(err)
	}

	fmt.Printf("host account's balance(before force close): %v\n\n", hostAccount.LoadBalance())
	return channelProposeMsg, channelAcceptMsg
}
