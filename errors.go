package main

import (
	"fmt"
	"github.com/stellar/go/clients/horizon"
)

func showDetailError(err error) {
	fmt.Println("detail error description:")

	err2 := err.(*horizon.Error).Problem
	tmpl := `
	Type:     %v
	Title:    %v
	Status:   %v
	Detail:   %v
	Instance: %v
	`
	fmt.Printf(tmpl, err2.Type, err2.Title, err2.Status, err2.Detail, err2.Instance)

	for key, value := range err2.Extras {
		fmt.Printf("key: %v, value: %v\n", key, string(value))
	}
}