// Package main contains the main CLI application.
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/andygello555/ginvoice/api"
	"github.com/andygello555/ginvoice/globals"
	"reflect"
	"strings"
	"time"
)

func main() {
	// Invoice number
	numberPtr := flag.Uint("number", 1, "The number of the invoice.")

	// From
	from := api.Contact{}
	flag.Var(&from, "from", "The contact who issued the invoice. (required, contact's Company defaults to their first and last name if not given)")

	// To
	to := api.Contact{}
	flag.Var(&to, "to", "The contact who needs to pay the invoice. (required, contact's Company defaults to their first and last name if not given)")

	// Bank
	bank := api.Bank{}
	flag.Var(&bank, "bank", "The bank details of the contact who issued the invoice. (optional)")

	// Date stuff
	invoiceDate := api.Date(time.Now())
	dueDate := api.Date(time.Now())
	flag.Var(&invoiceDate, "date", "The date the invoice was created.")
	flag.Var(&dueDate, "due", "The date on which the invoice needs to be paid.")

	// Invoice items
	items := make(api.Items, 0)
	flag.Var(&items, "items", "The items that the employee performed and needs to be paid for. (required, hrs/qty defaults to 1, tax defaults to 0)")

	// Parse
	flag.Parse()

	// Check the required flags
	required := map[string]interface{} {
		"from": &from,
		"to": &to,
		"items": &items,
	}
	needed := make([]string, 0)
	for flagName, req := range required {
		switch req.(type) {
		case *api.Contact:
			contact := req.(*api.Contact)
			if reflect.DeepEqual(*contact, api.Contact{}) {
				needed = append(needed, flagName)
			}
		case *api.Items:
			is := req.(*api.Items)
			if len(*is) == 0 {
				needed = append(needed, flagName)
			}
		}
	}
	if len(needed) != 0 {
		globals.RequiredFlag.Handle(errors.New(strings.Join(needed, ", ")))
	}

	fmt.Println("Invoice number:", *numberPtr)
	fmt.Println("From contact:", &from)
	fmt.Println("To contact:", &to)
	fmt.Println("Invoice date:", &invoiceDate)
	fmt.Println("Due date:", &dueDate)
	fmt.Println("Items:", &items)
}
