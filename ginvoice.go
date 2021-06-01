package main

import (
	"flag"
	"github.com/andygello555/ginvoice/api"
	"time"
)

func main() {
	// Invoice number
	flag.Uint("number", 0, "The number of the invoice.")

	// From
	from := api.Contact{}
	flag.Var(&from, "from", "The contact who issued the invoice (required, contact's Company defaults to their first and last name if not given)")

	// To
	to := api.Contact{}
	flag.Var(&to, "to", "The contact who needs to pay the invoice (required, contact's Company defaults to their first and last name if not given)")

	// Date stuff
	invoiceDate := api.Date(time.Now())
	dueDate := api.Date(time.Now())
	flag.Var(&invoiceDate, "date", "The date the invoice was created (default: time.Now())")
	flag.Var(&dueDate, "due", "The date on which the invoice needs to be paid (default: time.Now())")

	// Invoice items
	items := make(api.Items, 0)
	flag.Var(&items, "items", "The items that the employee performed and needs to be paid for (required, hrs/qty defaults to 1)")
}
