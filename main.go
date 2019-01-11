package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	unilateralClose := flag.Bool("unilateral_close", false, "")
	payment := flag.Bool("payment", false, "")
	htlcTimeoutPayment := flag.Bool("htlc_timeout_payment", false, "")
	htlcSuccessPayment := flag.Bool("htlc_success_payment", false, "")
	flag.Parse()

	if *htlcTimeoutPayment || *htlcSuccessPayment {
		htlcMode = true
	}

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
	if htlcMode {
		if err := hostAccount.setupAccountTx(htlcResolutionAccountType); err != nil {
			log.Fatal(err)
		}
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
		showDetailError(err)
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

		secsToWait := 2*defaultFinalityDelay + defaultMaxRoundDuration + 1
		explanation := "2*defaultFinalityDelay + defaultMaxRoundDuration + 1"
		fmt.Printf("waiting %v secs(%v) until settlement's txs will become valid\n", secsToWait, explanation)
		time.Sleep(time.Duration(secsToWait) * time.Second)

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

		if err := hostAccount.publishRatchetTx(paymentAcceptMsg.RecipientRatchetSig, paymentState); err != nil {
			log.Fatal(err)
		}

		secsToWait := 2*defaultFinalityDelay + defaultMaxRoundDuration + 1
		explanation := "2*defaultFinalityDelay + defaultMaxRoundDuration + 1"
		fmt.Printf("waiting %v secs(%v) until settlement's txs will become valid", secsToWait, explanation)
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

		return
	}

	if *htlcTimeoutPayment || *htlcSuccessPayment {
		rHash := guestAccount.addInvoice()
		fmt.Printf("hex-encoded rHash: %v\n", hex.EncodeToString(rHash[:]))

		paymentProposeMsg, err := hostAccount.createHTLCPaymentProposeMsg(1, guestAccount.keyPair.Address(), rHash)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(paymentProposeMsg)

		paymentAcceptMsg, err := guestAccount.receiveHTLCPaymentProposeMsg(paymentProposeMsg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(paymentAcceptMsg)

		if err := hostAccount.publishRatchetTx(paymentAcceptMsg.RecipientRatchetSig, htlcPaymentState); err != nil {
			log.Fatal(err)
		}

		secsToWait := 2*defaultFinalityDelay + defaultMaxRoundDuration + 1
		fmt.Printf("waiting %v secs until settlement's txs will become valid\n", secsToWait)
		time.Sleep(time.Duration(secsToWait) * time.Second)

		fmt.Println("createAndSignSettleOnlyWithHostAndActiveHtlcTx()")
		txe, err := hostAccount.createAndSignSettleOnlyWithHostAndActiveHtlcTx(
			uint64(rsn),
			paymentProposeMsg.PaymentTime,

			paymentAcceptMsg.RecipientSettleOnlyWithHostAndActiveHtlcSig,
			//uint64(rsn),
			//paymentProposeMsg.PaymentTime,
			//defaultPaymentAmount,
			//paymentAcceptMsg.RecipientSettleWithGuestSig,
		)
		if err != nil {
			log.Fatal(err)
		}

		if err := hostAccount.publishTx(txe); err != nil {
			showDetailError(err)
			log.Fatal(err)
		}

		if *htlcTimeoutPayment {
			htlcTimeoutTxe, err := hostAccount.createAndSignHtlcTimeoutTx(
				paymentAcceptMsg.RecipientHtlcTimeoutSig,
				paymentProposeMsg.PaymentTime,
			)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(loadBalance(hostAccount.htlcResolutionAccount.keyPair.Address()))
			fmt.Printf("sleep %v secs for htlcTimeout become valid\n", defaultFinalityDelay)
			time.Sleep(time.Second * defaultFinalityDelay)

			fmt.Println("publish htlcTimeoutTxe")
			if err := hostAccount.publishTx(htlcTimeoutTxe); err != nil {
				showDetailError(err)
				log.Fatal(err)
			}
		}

		if *htlcSuccessPayment {
			htlcSuccessTx, err := guestAccount.createAndSignHtlcSuccessTx(
				hostAccount.htlcResolutionAccount.keyPair.Address(),
				paymentProposeMsg.PaymentTime,
			)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(loadBalance(hostAccount.htlcResolutionAccount.keyPair.Address()))

			fmt.Println("publish htlcSuccessTx")
			if err := hostAccount.publishTx(htlcSuccessTx); err != nil {
				showDetailError(err)
				log.Fatal(err)
			}
		}

		fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.loadBalance())
		fmt.Printf("guest account's balance(after force close): %v\n\n", loadBalance(guestAccount.keyPair.Address()))

		return
	}
}
