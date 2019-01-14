package demo

import (
	"fmt"
	"log"
	"time"

	"github.com/LightningPeach/starlight-demo-impl/host"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/LightningPeach/starlight-demo-impl/wire"
)

func UnilateralClose(
	hostAccount *host.Account,
	channelProposeMsg *wire.ChannelProposeMsg,
	channelAcceptMsg *wire.ChannelAcceptMsg,
	rsn int,
) {
	fmt.Println("publish ratchetTx")
	tx, err := hostAccount.CreateAndSignRatchetTxForSelf(
		channelAcceptMsg.GuestRatchetRound1Sig,
		channelProposeMsg.FundingTime,
		rsn,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := hostAccount.PublishTx(tx); err != nil {
		log.Fatal(err)
	}

	secsToWait := 2*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration + 1
	explanation := "2*defaultFinalityDelay + defaultMaxRoundDuration + 1"
	fmt.Printf("waiting %v secs(%v) until settlement's txs will become valid\n", secsToWait, explanation)
	time.Sleep(time.Duration(secsToWait) * time.Second)

	fmt.Println("publish settleOnlyWithHostTx")
	if err := hostAccount.PublishSettleOnlyWithHostTx(channelAcceptMsg.GuestSettleOnlyWithHostSig, channelProposeMsg.FundingTime); err != nil {
		log.Fatal(err)
	}

	if hostAccount.LoadBalance() != "9999.9998600" {
		panic("err")
	}

	fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.LoadBalance())
}
