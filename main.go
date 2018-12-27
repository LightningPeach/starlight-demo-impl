package main

import (
	"fmt"
	//"github.com/stellar/go/build"
	//"github.com/stellar/go/clients/horizon"
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

	//paymentProposeMsg, err := hostAccount.createPaymentProposeMsg(1, guestAccount.keyPair.Address())
	//if err != nil {
	//	log.Fatal(err)
	//}

	//msg *PaymentProposeMsg,
	//escrowAddress,
	//hostRatchetAddress string,
	//bsn int64,
	//paymentAcceptMsg, err := guestAccount.receivePaymentProposeMsg(
	//	paymentProposeMsg,
	//	hostAccount.escrowKeyPair.Address(),
	//	hostAccount.hostRatchetAccount.keyPair.Address(),
	//	int64(hostAccount.baseSequenceNumber),
	//)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Println("RATCHET TX")
	//txCopy := paymentAcceptMsg.RecipientRatchetSig
	//txCopy.Mutate(build.Sign{Seed: hostAccount.escrowKeyPair.Seed()})
	//if err := hostAccount.publishTx(paymentAcceptMsg.RecipientRatchetSig); err != nil {
	//	fmt.Println("tx fail")
	//	err2 := err.(*horizon.Error).Problem
	//	fmt.Println("Type: ", err2.Type)
	//	fmt.Println("Title: ", err2.Title)
	//	fmt.Println("Status: ", err2.Status)
	//	fmt.Println("Detail:", err2.Detail)
	//	fmt.Println("Instance: ", err2.Instance)
	//	for key, value := range err2.Extras {
	//		fmt.Println("KEYVALUE: ", key, string(value))
	//	}
	//	// fmt.Println("Extras: ",   err2.Extras)
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("SETTLE WITH GUEST TX")
	//
	//fmt.Println("WAIT")
	//time.Sleep((2*defaultFinalityDelay + defaultMaxRoundDuration) * time.Second + 10 * time.Second)
	//
	//txCopy = paymentAcceptMsg.RecipientSettleWithGuestSig
	//// txCopy.Mutate(build.Sign{Seed:})
	//_ = txCopy
	//if err := hostAccount.publishTx(paymentAcceptMsg.RecipientSettleWithGuestSig); err != nil {
	//	fmt.Println("tx fail")
	//	err2 := err.(*horizon.Error).Problem
	//	fmt.Println("Type: ", err2.Type)
	//	fmt.Println("Title: ", err2.Title)
	//	fmt.Println("Status: ", err2.Status)
	//	fmt.Println("Detail:", err2.Detail)
	//	fmt.Println("Instance: ", err2.Instance)
	//	for key, value := range err2.Extras {
	//		fmt.Println("KEYVALUE: ", key, string(value))
	//	}
	//	// fmt.Println("Extras: ",   err2.Extras)
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("SETTLE WITH HOST TX")
	//txCopy = paymentAcceptMsg.RecipientSettleWithHostSig
	//fmt.Println(len(txCopy.E.Signatures))
	//fmt.Println(txCopy.E.Signatures)
	//_ = txCopy
	//if err := hostAccount.publishTx(paymentAcceptMsg.RecipientSettleWithHostSig); err != nil {
	//	fmt.Println("tx fail")
	//	err2 := err.(*horizon.Error).Problem
	//	fmt.Println("Type: ", err2.Type)
	//	fmt.Println("Title: ", err2.Title)
	//	fmt.Println("Status: ", err2.Status)
	//	fmt.Println("Detail:", err2.Detail)
	//	fmt.Println("Instance: ", err2.Instance)
	//	for key, value := range err2.Extras {
	//		fmt.Println("KEYVALUE: ", key, string(value))
	//	}
	//	// fmt.Println("Extras: ",   err2.Extras)
	//	log.Fatal(err)
	//}

	rsn := roundSequenceNumber(hostAccount.baseSequenceNumber, 1)

	fmt.Println("publish ratchetTx")
	tx, err := hostAccount.ratchetTx(
		channelAcceptMsg.GuestRatchetRound1Sig,
		channelProposeMsg.FundingTime,
		rsn,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := hostAccount.publishTx(tx); err != nil {
		log.Fatal(err)
	}

	time.Sleep((2*defaultFinalityDelay + defaultMaxRoundDuration) * time.Second + 10 * time.Second)
	fmt.Println("time.Now(): ", time.Now().Unix())

	fmt.Println("publish settleOnlyWithHostTx")
	if err := hostAccount.settleOnlyWithHostTx(channelAcceptMsg.GuestSettleOnlyWithHostSig); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.loadBalance())
	_ = channelAcceptMsg

	//fmt.Printf("guest account's balance(after force close): %v\n\n", loadBalance(guestAccount.keyPair.Address()))
	//_ = channelAcceptMsg
}
