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

	if err := hostAccount.fundingTx(guestAccount.keyPair.Address()); err != nil {
		log.Fatal(err)
	}
}
