package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("starlight_demo")

	fmt.Println("creating host account:")
	hostAccount, err := newHostAccount()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %v\n\n", hostAccount.loadBalance())

	fmt.Println("creating channel accounts:")
	fmt.Printf("creating: %v\n", hostRatchetAccount)
	if err := hostAccount.setupAccountTx(hostRatchetAccount); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %v\n\n", loadBalance(hostAccount.hostRatchetAccount.keyPair.Address()))

	fmt.Printf("creating: %v\n", guestRatchetAccount)
	if err := hostAccount.setupAccountTx(guestRatchetAccount); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %v\n\n", loadBalance(hostAccount.guestRatchetAccount.keyPair.Address()))

	fmt.Printf("creating: %v\n", escrowAccount)
	if err := hostAccount.setupAccountTx(escrowAccount); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %v\n\n", loadBalance(hostAccount.escrowKeyPair.Address()))

	fmt.Println("creating guest account:")
	guestAccount, err := newGuestAccount()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %v\n\n", loadBalance(guestAccount.keyPair.Address()))

	channelProposeMsg := hostAccount.createChannelProposeMsg(guestAccount.keyPair.Address())

	fmt.Println("receiveChannelProposeMsg: ")
	channelAcceptMsg, err := guestAccount.receiveChannelProposeMsg(channelProposeMsg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("publish fundingTx")
	if err := hostAccount.fundingTx(guestAccount.keyPair.Address()); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("host account's balance(before force close): %v\n\n", hostAccount.loadBalance())

	paymentProposeMsg, err := hostAccount.createPaymentProposeMsg(1, guestAccount.keyPair.Address())
	if err != nil {
		log.Fatal(err)
	}

	_ = paymentProposeMsg

	//fmt.Println("publish ratchetTx")
	//if err := hostAccount.ratchetTx(channelAcceptMsg.GuestRatchetRound1Sig); err != nil {
	//	log.Fatal(err)
	//}
	//
	//time.Sleep((2*defaultFinalityDelay + defaultMaxRoundDuration) * time.Second + 10 * time.Second)
	//fmt.Println("time.Now(): ", time.Now().Unix())
	//
	//fmt.Println("publish settleOnlyWithHostTx")
	//if err := hostAccount.settleOnlyWithHostTx(channelAcceptMsg.GuestSettleOnlyWithHostSig); err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.loadBalance())
	_ = channelAcceptMsg
}
