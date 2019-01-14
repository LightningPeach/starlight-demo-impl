package demo

import (
	"fmt"
	"log"
	"time"

	"github.com/LightningPeach/starlight-demo-impl/guest"
	"github.com/LightningPeach/starlight-demo-impl/host"
	"github.com/LightningPeach/starlight-demo-impl/tools"
)

func SimplePayment(hostAccount *host.Account, guestAccount *guest.Account, rsn int) {
	paymentProposeMsg, err := hostAccount.CreatePaymentProposeMsg(1, guestAccount.KeyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(paymentProposeMsg)

	paymentAcceptMsg, err := guestAccount.ReceivePaymentProposeMsg(paymentProposeMsg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(paymentAcceptMsg)

	if err := hostAccount.PublishRatchetTx(paymentAcceptMsg.RecipientRatchetSig, tools.PaymentState); err != nil {
		log.Fatal(err)
	}

	secsToWait := 2*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration + 1
	explanation := "2*defaultFinalityDelay + defaultMaxRoundDuration + 1"
	fmt.Printf("waiting %v secs(%v) until settlement's txs will become valid", secsToWait, explanation)
	time.Sleep(time.Duration(secsToWait) * time.Second)

	err = hostAccount.PublishSettleWithGuestTx(
		uint64(rsn),
		paymentProposeMsg.PaymentTime,
		tools.DefaultPaymentAmount,
		paymentAcceptMsg.RecipientSettleWithGuestSig,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = hostAccount.PublishSettleWithHostTx(
		uint64(rsn),
		paymentProposeMsg.PaymentTime,
		paymentAcceptMsg.RecipientSettleWithHostSig,
	)
	if err != nil {
		log.Fatal(err)
	}

	if hostAccount.LoadBalance() != "9899.9998500" {
		panic("err")
	}
	if tools.LoadBalance(guestAccount.KeyPair.Address()) != "10100.0000000" {
		panic("err")
	}

	fmt.Printf("host account's balance(after force close): %v\n\n", hostAccount.LoadBalance())
	fmt.Printf("guest account's balance(after force close): %v\n\n", tools.LoadBalance(guestAccount.KeyPair.Address()))
}
