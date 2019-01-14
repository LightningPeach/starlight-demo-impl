package main

import (
	"testing"

	"github.com/LightningPeach/starlight-demo-impl/demo"
	"github.com/LightningPeach/starlight-demo-impl/tools"
)

func TestUnilateralClose(t *testing.T) {
	hostAccount, guestAccount := demo.Setup()
	channelProposeMsg, channelAcceptMsg := demo.OpenChannel(hostAccount, guestAccount)
	rsn := tools.RoundSequenceNumber(hostAccount.BaseSequenceNumber, 1)

	demo.UnilateralClose(hostAccount, channelProposeMsg, channelAcceptMsg, rsn)
}

func TestSimplePayment(t *testing.T) {
	hostAccount, guestAccount := demo.Setup()
	demo.OpenChannel(hostAccount, guestAccount)
	rsn := tools.RoundSequenceNumber(hostAccount.BaseSequenceNumber, 1)

	demo.SimplePayment(hostAccount, guestAccount, rsn)
}

func TestHtlcPayment_Timeout(t *testing.T) {
	tools.HtlcMode = true
	hostAccount, guestAccount := demo.Setup()
	demo.OpenChannel(hostAccount, guestAccount)
	rsn := tools.RoundSequenceNumber(hostAccount.BaseSequenceNumber, 1)

	demo.HtlcPayment(hostAccount, guestAccount, rsn, true, false)
}

func TestHtlcPayment_Success(t *testing.T) {
	tools.HtlcMode = true
	hostAccount, guestAccount := demo.Setup()
	demo.OpenChannel(hostAccount, guestAccount)
	rsn := tools.RoundSequenceNumber(hostAccount.BaseSequenceNumber, 1)

	demo.HtlcPayment(hostAccount, guestAccount, rsn, false, true)
}