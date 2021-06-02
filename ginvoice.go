// Package main contains the main CLI application.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/andygello555/ginvoice/api"
	"github.com/andygello555/ginvoice/globals"
	"github.com/andygello555/gotils/files"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// We change flag.Usage to be a bit more explanatory on the types
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Printf(`
Custom types:

contact:
	Comma-seperated key-value pairs (seperated by "%s"):
		<key>: <value>, <key>: <value>, ...

	Possible keys (string literals in parenthesis denote key possibilities):
		- Company ("company", "comp", "c"): The name of the contact's company. (defaults to "FirstName + LastName")
		- FirstName ("firstname", "first", "f"): The first name of the contact. (required)
		- LastName ("lastname", "last", "l"): The last name of the contact. (required)
		- Email ("email", "e"): The email of the contact. (required and validated)
		- PhoneNo ("phoneno", "phone", "p"): The phone number of the contact. (required and validated)
		- Address ("address", "addr", "a"): The address of the of the contact. This is given as a "%s" seperated list. (required) 

items:
	Comma-seperated list of items where each item is a list of "%s" seperated key-value pairs (seperated by "%s"):
		<item>, <item>, ...
	Where <item> is:
		<key>: <value>%s <key>: <value>%s ...

	Possible keys (string literals in parenthesis denote key possibilities):
		- Description ("description", "desc", "d"): The description of the invoice item. (required)
		- HoursQuantity ("hoursquantity", "hours", "h"): The hours/quantity of the invoice item. (defaults to 1)
		- Rate ("rate", "r"), see money type: The rate charged for the invoice item. (required)
			- The currency is determined by the 3 letter currency abbreviation (e.g. USD/GBP) or symbol before the number:
				"GBP 10.00"
				"USD10.00"
				"£10.00"
		- Tax ("tax", "t"), see money type: The tax to be applied on top of the invoice item. (defaults to 0.00)

money:
	Money string used in items. The currency is determined by the 3 letter currency abbreviation (e.g. USD/GBP) or 
	symbol before the number. Here are some examples:
		"GBP 10.00"
		"USD10.00"
		"£10.00"

bank:
	Comma-seperated key-value pairs (seperated by "%s"):
		<key>: <value>, <key>: <value>, ...

	Possible keys (string literals in parenthesis denote key possibilities):
		- Bank ("bank", "b"): The name of the bank. (required)
		- Account No. ("accountno", "account", "a/c", "a"): The account number. (required)
		- Sort Code ("sortcode", "sort", "code", "s"): The sort code. (required)

date:
	Date in D/M/YYYY format (sorry Americans).
`, globals.KeyValueSep,
   globals.SecondLevelSep,
   globals.SecondLevelSep,
   globals.KeyValueSep,
   globals.SecondLevelSep,
   globals.SecondLevelSep,
   globals.KeyValueSep)
	}

	// Verbose
	verbosePtr := flag.Bool("verbose", false, "Whether or not to print some extra info.")

	// Invoice number
	numberPtr := flag.Uint("number", 1, "The number of the invoice.")

	// From
	from := api.Contact{}
	flag.Var(&from, "from", "The `contact` who issued the invoice. (required, contact's Company defaults to their first and last name if not given)")

	// To
	to := api.Contact{}
	flag.Var(&to, "to", "The `contact` who needs to pay the invoice. (required, contact's Company defaults to their first and last name if not given)")

	// Bank
	bank := api.Bank{}
	flag.Var(&bank, "bank", "The `bank` details of the contact who issued the invoice. (optional)")

	// Date stuff
	invoiceDate := api.Date(time.Now())
	dueDate := api.Date(time.Now())
	flag.Var(&invoiceDate, "date", "The `date` the invoice was created.")
	flag.Var(&dueDate, "due", "The `date` on which the invoice needs to be paid.")

	// Invoice items
	items := make(api.Items, 0)
	flag.Var(&items, "items", "The `items` that the employee performed and needs to be paid for. (required, hrs/qty defaults to 1, tax defaults to 0)")

	// Output file
	outputPathPtr := flag.String("output", "invoice.pdf", "The output filepath for the invoice.")

	// Parse
	flag.Parse()

	// Construct the invoice value to check required flags
	invoice, err := api.NewInvoice(*numberPtr, &from, &to, &items, &bank, &invoiceDate, &dueDate)
	if err != nil {
		globals.RequiredFlag.Handle(err)
	}

	// Print out the parsed information if verbose is given
	if *verbosePtr {
		fmt.Println("Parsed information:")
		fmt.Println(invoice)
	}

	// Generate the invoice
	var buf bytes.Buffer
	buf, err = invoice.Generate()

	if err != nil {
		globals.InvoiceGenerationErr.Handle(err)
	}

	// Save to file
	var f *os.File
	outputPathDir := filepath.Dir(*outputPathPtr)
	if !files.IsDir(outputPathDir) {
		globals.FileErrUser.Handle(errors.New(fmt.Sprintf("cannot write to directory: \"%s\", as it does not exist", *outputPathPtr)))
	}
	f, err = os.Create(*outputPathPtr)

	if err != nil {
		globals.FileErr.Handle(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			globals.FileErr.Handle(err)
		}
	}(f)

	_, err = buf.WriteTo(f)

	if err != nil {
		globals.FileErr.Handle(err)
	}
}
