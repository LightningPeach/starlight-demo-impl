package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("starlight_demo")

	fmt.Println("creating host account:")
	hostAccount, err := newHostAccount()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("creating channel accounts:")
	if err := hostAccount.setupAccountTx(hostRatchetAccount); err != nil {
		log.Fatal(err)
	}
	if err := hostAccount.setupAccountTx(guestRatchetAccount); err != nil {
		log.Fatal(err)
	}
	if err := hostAccount.setupAccountTx(escrowAccount); err != nil {
		log.Fatal(err)
	}

	fmt.Println("creating guest account:")
	guestAccount, err := newGuestAccount()
	if err != nil {
		log.Fatal(err)
	}

	channelProposeMsg := hostAccount.createChannelProposeMsg(guestAccount.keyPair.Address())

	channelAcceptMsg, err := guestAccount.receiveChannelProposeMsg(channelProposeMsg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("fundingTx")
	if err := hostAccount.fundingTx(guestAccount.keyPair.Address()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("ratchetTx")
	if err := hostAccount.ratchetTx(channelAcceptMsg.GuestRatchetRound1Sig); err != nil {
		log.Fatal(err)
	}

	time.Sleep((2*defaultFinalityDelay + defaultMaxRoundDuration) * time.Second)

	fmt.Println("settleOnlyWithHostTx")
	if err := hostAccount.settleOnlyWithHostTx(channelAcceptMsg.GuestSettleOnlyWithHostSig); err != nil {
		log.Fatal(err)
	}
}
