package api

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestContact(t *testing.T) {
	for _, test := range []struct{
		input string
		err   error
		out   Contact
	}{
		{
			input: "company: Company, firstName: John, lastName: Smith, email: johnsmith@example.com, phoneNo: 123123123, address: 1 Smith Street;Smith Town;Smith;SM20 123;UK",
			err:   nil,
			out:   Contact{
				Company:   "Company",
				FirstName: "John",
				LastName:  "Smith",
				Email:     "johnsmith@example.com",
				PhoneNo:   "123123123",
				Address:   []string{
					"1 Smith Street",
					"Smith Town",
					"Smith",
					"SM20 123",
					"UK",
				},
			},
		},
		{
			input: "firstName: John, lastName: Smith, email: johnsmith@example.com, phoneNo: 123123123, address: 1 Smith Street;Smith Town;Smith;SM20 123;UK",
			err:   nil,
			out:   Contact{
				Company:   "John Smith",
				FirstName: "John",
				LastName:  "Smith",
				Email:     "johnsmith@example.com",
				PhoneNo:   "123123123",
				Address:   []string{
					"1 Smith Street",
					"Smith Town",
					"Smith",
					"SM20 123",
					"UK",
				},
			},
		},
		{
			input: "lastName: Smith, email: johnsmith@example.com, phoneNo: 123123123, address: 1 Smith Street;Smith Town;Smith;SM20 123;UK",
			err:   errors.New("Contact details: you need to give 6 key-value pairs each of which representing one of the following fields:\n\t- Company\n\t- FirstName\n\t- LastName\n\t- Email\n\t- PhoneNo\n\t- Address\n"),
			out:   Contact{},
		},
		{
			input: "firstName; John, lastName: Smith, email: johnsmith@example.com, phoneNo: 123123123, address: 1 Smith Street;Smith Town;Smith;SM20 123;UK",
			err:   errors.New("Contact details: cannot find key in text: \"firstName; John\", keys must be one or more character long followed by a ':' then an optional whitespace"),
			out:   Contact{},
		},
		{
			input: "f:John,l:Smith,e:johnsmith@example.com,p:123123123,a:1 Smith Street;Smith Town;Smith;SM20 123;UK",
			err:   nil,
			out:   Contact{
				Company:   "John Smith",
				FirstName: "John",
				LastName:  "Smith",
				Email:     "johnsmith@example.com",
				PhoneNo:   "123123123",
				Address:   []string{
					"1 Smith Street",
					"Smith Town",
					"Smith",
					"SM20 123",
					"UK",
				},
			},
		},
		{
			input: "f:John,l:Smith,e:not an email,p:123123123,a:1 Smith Street;Smith Town;Smith;SM20 123;UK",
			err:   errors.New("\"not an email\" is not a valid email"),
			out:   Contact{},
		},
	} {
		var contact Contact
		err := contact.Set(test.input)
		if err != nil && test.err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("parsing contact \"%s\" returns the incorrect error:\nexpected: \"%s\"\ngot: \"%s\"", test.input, test.err.Error(), err.Error())
			}
		} else if err == nil && test.err != nil {
			t.Errorf("parsing contact \"%s\" does not return the expected error: \"%s\"", test.input, test.err.Error())
		} else if err != nil && test.err == nil {
			t.Errorf("parsing contact \"%s\" is not supposed to return error: \"%s\"", test.input, err.Error())
		} else {
			if !reflect.DeepEqual(contact, test.out) {
				t.Errorf("expected output (%v) does not match actual output: %v", test.out, contact)
			}
		}
	}
}

func TestBank(t *testing.T) {
	for _, test := range []struct{
		input string
		err   error
		out   Bank
	}{
		{
			input: "Bank: Bank O' Clock, AccountNo: 12312312, SortCode: 696969",
			out:   Bank{
				Bank:      "Bank O' Clock",
				AccountNo: "12312312",
				SortCode:  "696969",
			},
		},
		{
			input: "b:Bank O' Clock,a/c:12312312,s:696969",
			out:   Bank{
				Bank:      "Bank O' Clock",
				AccountNo: "12312312",
				SortCode:  "696969",
			},
		},
		{
			input: "b:Bank O' Clock,a/c:1231231,s:696969",
			err:   errors.New("\"1231231\" is not a valid sort code (8 digits)"),
			out:   Bank{},
		},
		{
			input: "b:Bank O' Clock,a/c:abcabcab,s:696969",
			err:   errors.New("\"abcabcab\" is not a valid sort code (not numeric)"),
			out:   Bank{},
		},
		{
			input: "b:Bank O' Clock,a/c:abcabc1,s:696969",
			err:   errors.New("\"abcabc1\" is not a valid sort code (8 digits) (not numeric)"),
			out:   Bank{},
		},
		{
			input: "b:Bank O' Clock,a/c:12312312,s:69696",
			err:   errors.New("\"69696\" is not a valid sort code (6 digits)"),
			out:   Bank{},
		},
		{
			input: "b:Bank O' Clock,a/c:12312312,s:ababab",
			err:   errors.New("\"ababab\" is not a valid sort code (not numeric)"),
			out:   Bank{},
		},
		{
			input: "b:Bank O' Clock,a/c:12312312,s:abab1",
			err:   errors.New("\"abab1\" is not a valid sort code (6 digits) (not numeric)"),
			out:   Bank{},
		},
	} {
		var bank Bank
		err := bank.Set(test.input)
		if err != nil && test.err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("parsing bank \"%s\" returns the incorrect error:\nexpected: \"%s\"\ngot: \"%s\"", test.input, test.err.Error(), err.Error())
			}
		} else if err == nil && test.err != nil {
			t.Errorf("parsing bank \"%s\" does not return the expected error: \"%s\"", test.input, test.err.Error())
		} else if err != nil && test.err == nil {
			t.Errorf("parsing bank \"%s\" is not supposed to return error: \"%s\"", test.input, err.Error())
		} else {
			if !reflect.DeepEqual(bank, test.out) {
				t.Errorf("expected output (%v) does not match actual output: %v", test.out, bank)
			}
		}
	}
}

func TestItems(t *testing.T) {
	for _, test := range []struct{
		input     string
		err       error
		out       Items
		subtotals []Money
		total     Money
	}{
		{
			input: "d: Did thing 1; h: 10; r: $10; t: $0.1,d: Did thing 2; h: 10; r: $10; t: $0.1",
			out:   Items{
				{
					"Did thing 1",
					10,
					Money{
						Money:    1000,
						Currency: UnitedStatesDollar,
					},
					Money{
						Money:    10,
						Currency: UnitedStatesDollar,
					},
				},
				{
					"Did thing 2",
					10,
					Money{
						Money:    1000,
						Currency: UnitedStatesDollar,
					},
					Money{
						Money:    10,
						Currency: UnitedStatesDollar,
					},
				},
			},
			subtotals: []Money{
				{
					Money:    10010,
					Currency: UnitedStatesDollar,
				},
				{
					Money:    10010,
					Currency: UnitedStatesDollar,
				},
			},
			total: Money{
				20020,
				UnitedStatesDollar,
			},
		},
		{
			input: "d: Did thing 1; h: 10; r: $10,d: Did thing 2; h: 10; r: $10; t: $0.1",
			out:   Items{
				{
					"Did thing 1",
					10,
					Money{
						Money:    1000,
						Currency: UnitedStatesDollar,
					},
					Money{
						Money:    0,
						Currency: ZeroCurrency,
					},
				},
				{
					"Did thing 2",
					10,
					Money{
						Money:    1000,
						Currency: UnitedStatesDollar,
					},
					Money{
						Money:    10,
						Currency: UnitedStatesDollar,
					},
				},
			},
			subtotals: []Money{
				{
					Money:    10000,
					Currency: UnitedStatesDollar,
				},
				{
					Money:    10010,
					Currency: UnitedStatesDollar,
				},
			},
			total: Money{
				20010,
				UnitedStatesDollar,
			},
		},
		{
			input: "d:Did thing 1;r:$10",
			out:   Items{
				{
					"Did thing 1",
					1,
					Money{
						Money:    1000,
						Currency: UnitedStatesDollar,
					},
					Money{
						Money:    0,
						Currency: ZeroCurrency,
					},
				},
			},
			subtotals: []Money{
				{
					Money:    1000,
					Currency: UnitedStatesDollar,
				},
			},
			total: Money{
				1000,
				UnitedStatesDollar,
			},
		},
		{
			input:     "r:$10",
			err: 	   errors.New(`Item details: you need to give 4 key-value pairs each of which representing one of the following fields:
	- Description
	- HoursQuantity
	- Rate
	- Tax`),
			out:       Items{},
			subtotals: []Money{},
			total:     Money{},
		},
	} {
		items := make(Items, 0)
		err := items.Set(test.input)
		if err != nil && test.err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("parsing items \"%s\" returns the incorrect error:\nexpected: \"%s\"\ngot: \"%s\"", test.input, test.err.Error(), err.Error())
			}
		} else if err == nil && test.err != nil {
			t.Errorf("parsing items \"%s\" does not return the expected error: \"%s\"", test.input, test.err.Error())
		} else if err != nil && test.err == nil {
			t.Errorf("parsing items \"%s\" is not supposed to return error: \"%s\"", test.input, err.Error())
		} else {
			if !reflect.DeepEqual(items, test.out) {
				t.Errorf("expected output (%v) does not match actual output: %v", test.out, items)
			}

			// Check item subtotals
			for i, item := range items {
				if *item.Subtotal() != test.subtotals[i] {
					t.Errorf("item %d does not have the expected subtotal of %v, instead it has: %v", i + 1, &test.subtotals[i], item.Subtotal())
				}
			}

			// Check items total
			if *items.Total() != test.total {
				t.Errorf("items do not have the expected total of %v, instead it is: %v", &test.total, items.Total())
			}
		}
	}
}

func TestDate(t *testing.T) {
	for _, test := range []struct{
		input string
		err   error
		out   Date
		str   string
	}{
		{
			input: "10/12/2021",
			out:   Date(time.Date(2021, time.December, 10, 0, 0, 0, 0, time.UTC)),
			str:   "December 10th, 2021",
		},
		{
			input: "21/13/2000",
			err:   errors.New("parsing time \"21/13/2000\": month out of range"),
			out:   Date{},
		},
	} {
		var date Date
		err := date.Set(test.input)
		if err != nil && test.err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("parsing date \"%s\" returns the incorrect error:\nexpected: \"%s\"\ngot: \"%s\"", test.input, test.err.Error(), err.Error())
			}
		} else if err == nil && test.err != nil {
			t.Errorf("parsing date \"%s\" does not return the expected error: \"%s\"", test.input, test.err.Error())
		} else if err != nil && test.err == nil {
			t.Errorf("parsing date \"%s\" is not supposed to return error: \"%s\"", test.input, err.Error())
		} else {
			// Check equality
			if date != test.out {
				t.Errorf("expected output (%v) does not match actual output: %v", test.out, date)
			}

			// Check string output
			if date.String() != test.str {
				t.Errorf("expected String() output (%v) does not match actual output: %v", test.str, date.String())
			}
		}
	}
}
