package main

import (
	"flag"
	"fmt"
	//"github.com/stellar/go/build"
	//"github.com/stellar/go/clients/horizon"
	"log"
	"time"
)

func main() {
	unilateralClose := flag.Bool("unilateral_close", false, "")
	payment := flag.Bool("payment", false, "")
	flag.Parse()

	fmt.Println("starlight_demo")

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

	guestAccount, err := newGuestAccount()
	if err != nil {
		log.Fatal(err)
	}

	channelProposeMsg := hostAccount.createChannelProposeMsg(guestAccount.keyPair.Address())
	fmt.Println(channelProposeMsg)

	channelAcceptMsg, err := guestAccount.receiveChannelProposeMsg(channelProposeMsg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(channelAcceptMsg)

	if err := hostAccount.publishFundingTx(guestAccount.keyPair.Address()); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("host account's balance(before force close): %v\n\n", hostAccount.loadBalance())

	rsn := roundSequenceNumber(hostAccount.baseSequenceNumber, 1)

	if *unilateralClose {
		fmt.Println("publish ratchetTx")
		tx, err := hostAccount.createAndSignRatchetTxForSelf(
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
		if err := hostAccount.settleOnlyWithHostTx(channelAcceptMsg.GuestSettleOnlyWithHostSig, channelProposeMsg.FundingTime); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.loadBalance())

		return
	}

	if *payment {
		paymentProposeMsg, err := hostAccount.createPaymentProposeMsg(1, guestAccount.keyPair.Address())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(paymentProposeMsg)

		paymentAcceptMsg, err := guestAccount.receivePaymentProposeMsg(paymentProposeMsg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(paymentAcceptMsg)

		if err := hostAccount.publishRatchetTx(paymentAcceptMsg.RecipientRatchetSig); err != nil {
			log.Fatal(err)
		}


		secsToWait := 2*defaultFinalityDelay + defaultMaxRoundDuration + 10
		fmt.Printf("waiting %v secs until settlement's txs will become valid", secsToWait)
		time.Sleep(time.Duration(secsToWait) * time.Second)

		err = hostAccount.publishSettleWithGuestTx(
			uint64(rsn),
			paymentProposeMsg.PaymentTime,
			defaultPaymentAmount,
			paymentAcceptMsg.RecipientSettleWithGuestSig,
		)
		if err != nil {
			log.Fatal(err)
		}

		err = hostAccount.publishSignSettleWithHostTx(
			uint64(rsn),
			paymentProposeMsg.PaymentTime,
			paymentAcceptMsg.RecipientSettleWithHostSig,
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.loadBalance())
		fmt.Printf("guest account's balance(after force close): %v\n\n", loadBalance(guestAccount.keyPair.Address()))
	}
}
