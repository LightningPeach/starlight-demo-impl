package demo

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/LightningPeach/starlight-demo-impl/guest"
	"github.com/LightningPeach/starlight-demo-impl/host"
	"github.com/LightningPeach/starlight-demo-impl/tools"
)

func HtlcPayment(hostAccount *host.Account, guestAccount *guest.Account, rsn int, htlcTimeoutPayment bool, htlcSuccessPayment bool) {
	rHash := guestAccount.AddInvoice()
	fmt.Printf("hex-encoded rHash: %v\n", hex.EncodeToString(rHash[:]))

	paymentProposeMsg, err := hostAccount.CreateHTLCPaymentProposeMsg(1, guestAccount.KeyPair.Address(), rHash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(paymentProposeMsg)

	paymentAcceptMsg, err := guestAccount.ReceiveHTLCPaymentProposeMsg(paymentProposeMsg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(paymentAcceptMsg)

	if err := hostAccount.PublishRatchetTx(paymentAcceptMsg.RecipientRatchetSig, tools.HtlcPaymentState); err != nil {
		log.Fatal(err)
	}

	secsToWait := 2*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration + 1
	fmt.Printf("waiting %v secs until settlement's txs will become valid\n", secsToWait)
	time.Sleep(time.Duration(secsToWait) * time.Second)

	fmt.Println("createAndSignSettleOnlyWithHostAndActiveHtlcTx()")
	txe, err := hostAccount.CreateAndSignSettleOnlyWithHostAndActiveHtlcTx(
		uint64(rsn),
		paymentProposeMsg.PaymentTime,

		paymentAcceptMsg.RecipientSettleOnlyWithHostAndActiveHtlcSig,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := hostAccount.PublishTx(txe); err != nil {
		tools.ShowDetailError(err)
		log.Fatal(err)
	}

	if htlcTimeoutPayment {
		htlcTimeoutTxe, err := hostAccount.CreateAndSignHtlcTimeoutTx(
			paymentAcceptMsg.RecipientHtlcTimeoutSig,
			paymentProposeMsg.PaymentTime,
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(tools.LoadBalance(hostAccount.HtlcResolutionAccount.KeyPair.Address()))
		fmt.Printf("sleep %v secs for htlcTimeout become valid\n", tools.DefaultFinalityDelay)
		time.Sleep(time.Second * tools.DefaultFinalityDelay)

		fmt.Println("publish htlcTimeoutTxe")
		if err := hostAccount.PublishTx(htlcTimeoutTxe); err != nil {
			tools.ShowDetailError(err)
			log.Fatal(err)
		}

		if hostAccount.LoadBalance() != "9999.9997800" {
			panic("err")
		}
		if tools.LoadBalance(guestAccount.KeyPair.Address()) != "10000.0000000" {
			panic("err")
		}
	}

	if htlcSuccessPayment {
		htlcSuccessTx, err := guestAccount.CreateAndSignHtlcSuccessTx(
			hostAccount.HtlcResolutionAccount.KeyPair.Address(),
			paymentProposeMsg.PaymentTime,
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(tools.LoadBalance(hostAccount.HtlcResolutionAccount.KeyPair.Address()))

		fmt.Println("publish htlcSuccessTx")
		if err := hostAccount.PublishTx(htlcSuccessTx); err != nil {
			tools.ShowDetailError(err)
			log.Fatal(err)
		}

		if hostAccount.LoadBalance() != "9897.4997800" {
			panic("err")
		}
		if tools.LoadBalance(guestAccount.KeyPair.Address()) != "10100.0000000" {
			panic("err")
		}
	}

	fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.LoadBalance())
	fmt.Printf("guest account's balance(after force close): %v\n\n", tools.LoadBalance(guestAccount.KeyPair.Address()))
}
