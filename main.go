package main

import (
	"flag"

	"github.com/LightningPeach/starlight-demo-impl/demo"
	"github.com/LightningPeach/starlight-demo-impl/tools"
)

func main() {
	unilateralClose := flag.Bool(
		"unilateral_close",
		false,
		"setup -> open_channel -> unilateral close(by host); on-chain balance should not change",
	)
	payment := flag.Bool(
		"payment",
		false,
		"setup -> open_channel -> simple_payment(host -> guest) -> unilateral close(by host); " +
			"on-chain balance should change",
	)
	htlcTimeoutPayment := flag.Bool(
		"htlc_timeout_payment",
		false,
		"setup -> open_channel -> htlc_payment(host -> guest, fail with timeout) -> unilateral close with active htlc(by host);" +
			"on-chain balance should not change",
	)
	htlcSuccessPayment := flag.Bool(
		"htlc_success_payment",
		false,
		"setup -> open_channel -> htlc_payment(host -> guest, success) -> unilateral close with active htlc(by host);" +
			"on-chain balance should change",
	)
	flag.Parse()

	if *htlcTimeoutPayment || *htlcSuccessPayment {
		tools.HtlcMode = true
	}

	hostAccount, guestAccount := demo.Setup()
	channelProposeMsg, channelAcceptMsg := demo.OpenChannel(hostAccount, guestAccount)
	rsn := tools.RoundSequenceNumber(hostAccount.BaseSequenceNumber, 1)

	if *unilateralClose {
		demo.UnilateralClose(hostAccount, channelProposeMsg, channelAcceptMsg, rsn)
		return
	}

	if *payment {
		demo.SimplePayment(hostAccount, guestAccount, rsn)
		return
	}

	if *htlcTimeoutPayment || *htlcSuccessPayment {
		demo.HtlcPayment(hostAccount, guestAccount, rsn, *htlcTimeoutPayment, *htlcSuccessPayment)
		return
	}
}
