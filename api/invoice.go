package api

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/andygello555/gotils/ints"
	str "github.com/andygello555/gotils/strings"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"reflect"
	"strconv"
	"strings"
)

type Invoice struct {
	Number      uint
	From        *Contact
	To          *Contact
	Items       *Items
	Bank        *Bank
	InvoiceDate *Date
	DueDate     *Date
}

// NewInvoice constructs a new invoice From the given flags and checks if the required flags are not empty.
func NewInvoice(number uint, from, to *Contact, items *Items, bank *Bank, invoiceDate, dueDate *Date) (*Invoice, error) {
	i := Invoice{
		Number:      number,
		From:        from,
		To:          to,
		Items:       items,
		Bank:        bank,
		InvoiceDate: invoiceDate,
		DueDate:     dueDate,
	}

	// Check required fields
	needed := make([]string, 0)
	addToNeeded := func(fieldName string) {
		needed = append(needed, fieldName)
	}
	for _, reqField := range []string{"From", "To", "Items"} {
		req := reflect.Indirect(reflect.ValueOf(i)).FieldByName(reqField).Interface()
		switch req.(type) {
		case *Contact:
			contact := req.(*Contact)
			if reflect.DeepEqual(*contact, Contact{}) {
				addToNeeded(reqField)
			}
		case *Items:
			is := req.(*Items)
			if len(*is) == 0 {
				addToNeeded(reqField)
			}
		}
	}
	if len(needed) > 0 {
		return nil, errors.New(strings.Join(needed, ", "))
	}
	return &i, nil
}

func (i *Invoice) Generate() (bytes.Buffer, error) {
	invoiceNumber := fmt.Sprintf("%03d", i.Number)
	darkGrayColor := getDarkGrayColor()
	grayColor := getGrayColor()
	lightGrayColor := getLightGrayColor()
	whiteColor := color.NewWhite()
	header := getHeader()
	contents := i.getContents()

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(10, 15, 10)

	contactTextProps := props.Text{
		Top:   3,
		Style: consts.Normal,
		Size:  9,
		Align: consts.Left,
		Color: darkGrayColor,
	}

	emptyClosure := func() {}

	m.RegisterHeader(func() {
		// INVOICE title
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text("INVOICE " + invoiceNumber, props.Text{
					Top:             3,
					Style:           consts.Bold,
					Size:            14,
					Align:           consts.Center,
				})
			})
		})

		m.Row(5, emptyClosure)

		// From and To headers
		m.Row(5, func() {
			m.Col(4, func() {
				m.Text("FROM", props.Text{
					Top:   3,
					Style: consts.Bold,
					Size:  8,
					Align: consts.Left,
					Color: grayColor,
				})
			})
			m.Col(4, func() {
				m.Text("TO", props.Text{
					Top:   3,
					Style: consts.Bold,
					Size:  8,
					Align: consts.Left,
					Color: grayColor,
				})
			})
		})

		m.Row(5, emptyClosure)

		// Companies
		m.Row(5, func() {
			m.Col(4, func() {
				m.Text(i.From.Company, props.Text{
					Top:   3,
					Style: consts.Bold,
					Size:  9,
					Align: consts.Left,
				})
			})
			m.Col(4, func() {
				m.Text(i.To.Company, props.Text{
					Top:   3,
					Style: consts.Bold,
					Size:  9,
					Align: consts.Left,
				})
			})
		})

		// First and last names
		m.Row(5, func() {
			m.Col(4, func() {
				m.Text(i.From.FirstName + " " + i.From.LastName, contactTextProps)
			})
			m.Col(4, func() {
				m.Text(i.To.FirstName + " " + i.To.LastName, contactTextProps)
			})
		})

		// Addresses
		addresses := ints.Max(len(i.From.Address), len(i.To.Address))
		fromAddressLen := len(i.From.Address)
		toAddressLen := len(i.To.Address)
		for a := 0; a < addresses; a++ {
			m.Row(5, func() {
				if a < fromAddressLen {
					m.Col(4, func() {
						m.Text(i.From.Address[a], contactTextProps)
					})
				} else {
					m.ColSpace(4)
				}
				if a < toAddressLen {
					m.Col(4, func() {
						m.Text(i.To.Address[a], contactTextProps)
					})
				} else {
					m.ColSpace(4)
				}
			})
		}

		m.Row(2, emptyClosure)

		// Email
		m.Row(5, func() {
			m.Col(4, func() {
				m.Text(i.From.Email, contactTextProps)
			})
			m.Col(4, func() {
				m.Text(i.To.Email, contactTextProps)
			})
		})

		// Phone numbers
		m.Row(5, func() {
			m.Col(4, func() {
				m.Text(i.From.PhoneNo, contactTextProps)
			})
			m.Col(4, func() {
				m.Text(i.To.PhoneNo, contactTextProps)
			})
		})
	})

	// If bank details are given then add those in
	if *i.Bank != (Bank{}) {
		m.Row(5, emptyClosure)
		quickString := func(s string) {
			m.Row(5, func() {
				m.Col(4, func() {
					m.Text(s, contactTextProps)
				})
			})
		}
		quickString("Bank details:")
		quickString(i.Bank.Bank)
		quickString("A/c No.     " + i.Bank.AccountNo)
		quickString("Sort code: " + i.Bank.SortCode)
	}

	m.Row(6, emptyClosure)

	m.Row(6, func() {
		m.Col(2, func() {
			m.Text("Invoice No.:", props.Text{
				Top:   0,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(3, func() {
			m.Text(invoiceNumber, props.Text{
				Top:   0,
				Style: consts.Normal,
				Align: consts.Left,
			})
		})
		m.ColSpace(7)
	})

	m.Row(6, func() {
		m.Col(2, func() {
			m.Text("Invoice Date:", props.Text{
				Top:   0,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(3, func() {
			m.Text(i.InvoiceDate.String(), props.Text{
				Top:   0,
				Style: consts.Normal,
				Align: consts.Left,
			})
		})
		m.ColSpace(3)
		m.Col(2, func() {
			m.Text("Due:", props.Text{
				Top:   0,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(2, func() {
			m.Text(i.DueDate.String(), props.Text{
				Top:   0,
				Style: consts.Normal,
				Align: consts.Left,
			})
		})
	})

	m.Row(7, emptyClosure)

	// Item rundown
	m.SetBackgroundColor(lightGrayColor)
	m.TableList(header, contents, props.TableList{
		HeaderProp: props.TableListContent{
			Size:      10.5,
			GridSizes: []uint{4, 2, 2, 2, 2},
		},
		ContentProp: props.TableListContent{
			Size:      11,
			GridSizes: []uint{4, 2, 2, 2, 2},
		},
		Align:                consts.Center,
		AlternatedBackground: &grayColor,
		HeaderContentSpace:   1,
		Line:                 false,
	})

	// Total
	m.RegisterFooter(func() {
		m.Row(10, emptyClosure)
		m.Row(10, func() {
			m.ColSpace(8)
			m.SetBackgroundColor(grayColor)
			m.Col(4, func() {
				m.Text("Invoice Summary", props.Text{
					Top:   3,
					Size:  11,
					Style: consts.Bold,
					Align: consts.Center,
					Color: darkGrayColor,
				})
			})
		})
		m.Row(10, func() {
			m.SetBackgroundColor(whiteColor)
			m.ColSpace(8)
			m.SetBackgroundColor(lightGrayColor)
			m.Col(2, func() {
				m.Text("Total: ", props.Text{
					Top:   3,
					Style: consts.Bold,
					Size:  9,
					Align: consts.Right,
				})
			})
			m.Col(2, func() {
				m.Text(i.Items.Total().StringAbbr(), props.Text{
					Top:   3,
					Style: consts.Bold,
					Size:  9,
					Align: consts.Left,
				})
			})
		})
	})

	buf, err := m.Output()
	return buf, err
}

func getHeader() []string {
	itemType := reflect.TypeOf(Item{})
	headers := make([]string, 0)
	for i := 0; i < itemType.NumField(); i++ {
		headers = append(headers, str.JoinCamelcase(itemType.Field(i).Name, "/"))
	}
	headers = append(headers, "Subtotal")
	return headers
}

func (i *Invoice) getContents() [][]string {
	contents := make([][]string, 0)
	for _, item := range *i.Items {
		hrsQty := strconv.Itoa(int(item.HoursQuantity))
		rate := fmt.Sprintf("%.2f", item.Rate.Float64())
		tax := item.Tax.StringAbbr()
		contents = append(contents, []string{item.Description, hrsQty, rate, tax, item.Subtotal().StringAbbr()})
	}
	return contents
}

func getDarkGrayColor() color.Color {
	return color.Color{
		Red:   55,
		Green: 55,
		Blue:  55,
	}
}

func getGrayColor() color.Color {
	return color.Color{
		Red:   200,
		Green: 200,
		Blue:  200,
	}
}

func getLightGrayColor() color.Color {
	return color.Color{
		Red:   240,
		Green: 240,
		Blue:  240,
	}
}

func (i *Invoice) String() string {
	return fmt.Sprintf(`Invoice Number: %d

From contact: %v
To contact: %v

Invoice date: %v
Due date: %v

Items: %v`, i.Number, i.From, i.To, i.InvoiceDate, i.DueDate, i.Items)
}
