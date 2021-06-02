// Package api contains all the background functions, types and constants To parse command line arguments and create invoices.
package api

import (
	"errors"
	"fmt"
	"github.com/andygello555/ginvoice/globals"
	"github.com/andygello555/gotils/ints"
	"github.com/andygello555/gotils/misc"
	str "github.com/andygello555/gotils/strings"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type KeyValueFlags interface {
	KeyVal(keyVal string) (interface{}, error)
}

func keyValLogic(keyVal string, t KeyValueFlags, secondLevelSplit *regexp.Regexp, possibleKeyMapping *map[interface{}]map[string]struct{}) (interface{}, error) {
	found := false
	keyValSplit := globals.KeyValueSplit.Split(keyVal, 2)
	keyValName := strings.Trim(str.TypeName(t), "*api.")

	// If we don't find a key and a value then we fill out an appropriate error
	if len(keyValSplit) != 2 {
		// Create a nice help text lookup
		var builder strings.Builder
		i := 0
		for _, mapping := range *possibleKeyMapping {
			builder.WriteString("\t- The following are all possible prefixes for \"" + reflect.TypeOf(t).Elem().Field(i).Name + "\"\n")
			for possible := range mapping {
				builder.WriteString("\t\t- " + possible + "\n")
			}
			i ++
		}
		builder.WriteString("\n")
		return nil, errors.New(fmt.Sprintf("%s details: cannot find key in text: \"%s\", keys must be one or more character long followed by a ':' then an optional whitespace\nany one of the following keys are valid (case insensitive):\n%s", keyValName, keyVal, builder.String()))
	}

	key := strings.ToLower(keyValSplit[0])
	val := keyValSplit[1]

	// We iterate over all the possible mappings and set the relative fields accordingly
	var prop interface{}
	var possible map[string]struct{}
	for prop, possible = range *possibleKeyMapping {
		if _, ok := possible[key]; ok {
			found = true
			switch prop.(type) {
			case *string:
				*prop.(*string) = val
			case *[]string:
				// If we are on a string slice then we will split at each second level char
				*prop.(*[]string) = secondLevelSplit.Split(val, -1)
			case *uint:
				i, err := strconv.Atoi(val)
				if err != nil {
					return nil, err
				}
				*prop.(*uint) = uint(i)
			case *Money:
				m, err := ParseMoney(val)
				if err != nil {
					return nil, err
				}
				*prop.(*Money) = *m
			default:
				return nil, errors.New(fmt.Sprintf("cannot set field of type \"%s\" To \"%v\"", str.TypeName(prop), val))
			}
			break
		}
	}

	// If we didn't find any key-value pairs then we return an appropriate error
	var err error = nil
	if !found {
		err = errors.New(fmt.Sprintf("%s details: could not find any key-value pair within \"%s\"", keyValName, keyVal))
	}
	return prop, err
}

func setLogic(value string, t KeyValueFlags, firstLevelSplit *regexp.Regexp) error {
	typeName := strings.TrimLeft(str.TypeName(t), "*api.")
	var err error = nil
	var fieldP interface{}
	foundKeysSet := make(map[interface{}]struct{})
	for _, keyVal := range firstLevelSplit.Split(value, -1) {
		if fieldP, err = t.KeyVal(keyVal); err == nil {
			if _, ok := foundKeysSet[fieldP]; !ok {
				foundKeysSet[fieldP] = struct{}{}
			} else {
				err = errors.New(fmt.Sprintf("specified the same %s field multiple times in list of key-value pairs", typeName))
				break
			}
			continue
		}
		break
	}

	// If we have a mismatch of the Number of fields then fill out an error
	bankType := reflect.TypeOf(t)
	numFields := bankType.Elem().NumField()
	if err == nil && numFields != len(foundKeysSet) {
		// Here we append a list of all fields within the Bank struct
		var fieldNames strings.Builder
		for i := 0; i < numFields; i++ {
			fieldNames.WriteString("\t- " + bankType.Elem().Field(i).Name + "\n")
		}
		err = errors.New(fmt.Sprintf("%s details: you need To give %d key-value pairs each of which representing one of the following fields:\n%s", typeName, numFields, fieldNames.String()))
	}
	return err
}

type Item struct {
	Description   string
	HoursQuantity uint
	Rate          Money
	Tax           Money
}

func (i *Item) Subtotal() *Money {
	return i.Rate.Multiply(float64(i.HoursQuantity)).Add(i.Tax.Float64())
}

func (i *Item) String() string {
	return fmt.Sprintf("HRS/QTY: %d, RATE: %s, TAX: %s, Subtotal: %s", i.HoursQuantity, i.Rate.String(), i.Tax.String(), i.Subtotal().String())
}

func (i *Item) KeyVal(keyVal string) (interface{}, error) {
	possibleKeyMappings := map[interface{}]map[string]struct{}{
		&i.Description: {
			"description": {},
			"desc": {},
			"d": {},
		},
		&i.HoursQuantity: {
			"hoursquantity": {},
			"hours":         {},
			"hrs":           {},
			"h":             {},
			"quantity":      {},
			"qty":           {},
			"q":             {},
		},
		&i.Rate: {
			"rate": {},
			"r":    {},
		},
		&i.Tax: {
			"tax": {},
			"t":   {},
		},
	}
	return keyValLogic(keyVal, i, globals.ThirdLevelSplit, &possibleKeyMappings)
}

type Items []*Item

func (is *Items) Total() *Money {
	var currency Currency
	if len(*is) > 0 {
		currency = (*is)[0].Rate.Currency
	} else {
		currency = ZeroCurrency
	}
	total := ToMoney(0, currency)
	for _, item := range *is {
		total = total.Add(item.Subtotal().Float64())
	}
	return total
}

func (is *Items) String() string {
	var b strings.Builder
	for i, item := range *is {
		b.WriteString(fmt.Sprintf("Item %d: %s", i + 1, item.String()))
	}
	return b.String()
}

func (is *Items) Set(value string) error {
	var err error = nil
	items := make([]*Item, 0)
	for _, itemStr := range globals.FirstLevelSplit.Split(value, -1) {
		item := Item{}
		err = setLogic(itemStr, &item, globals.SecondLevelSplit)
		// Here we count how many empty fields we have using reflection. This is so we can default the HoursQuantity
		// value To 1 if no HoursQuantity value is given.
		valueOf := reflect.ValueOf(item)
		emptyMoney := Money{}
		clear := true
		for i := 0; i < reflect.TypeOf(item).NumField(); i++ {
			val := reflect.Indirect(valueOf).Field(i)
			switch reflect.TypeOf(item).Field(i).Name {
			case "HoursQuantity":
				if val.Interface().(uint) == 0 {
					item.HoursQuantity = 1
				}
			case "Tax":
				if val.Interface().(Money) == emptyMoney {
					item.Tax = Money{
						Money:    0,
						Currency: ZeroCurrency,
					}
				}
			case "Description":
				if len(val.String()) == 0 {
					clear = false
				}
			case "Rate":
				if val.Interface().(Money) == emptyMoney {
					clear = false
				}
			}
		}

		// Here clear is true then we can clear out the error
		if clear {
			err = nil
		}
		if err != nil {
			return err
		}
		items = append(items, &item)
	}
	*is = items
	return err
}

type Contact struct {
	Company   string
	FirstName string
	LastName  string
	Email     string
	PhoneNo   string
	Address   []string
}

func (c *Contact) String() string {
	return fmt.Sprintf(`%s
%s %s
%s

%s
%s
`, c.Company, c.FirstName, c.LastName, strings.Join(c.Address, "\n"), c.Email, c.PhoneNo)
}

func (c *Contact) KeyVal(keyVal string) (interface{}, error) {
	possibleKeyMappings := map[interface{}]map[string]struct{} {
		&c.Company: {
			"company": {},
			"c": {},
			"comp": {},
		},
		&c.FirstName: {
			"firstname": {},
			"first": {},
			"f": {},
		},
		&c.LastName: {
			"lastname": {},
			"last": {},
			"l": {},
		},
		&c.Email: {
			"email": {},
			"e": {},
		},
		&c.PhoneNo: {
			"phoneno": {},
			"phonenumber": {},
			"phone": {},
			"p": {},
		},
		&c.Address: {
			"address": {},
			"addr": {},
			"a": {},
		},
	}
	return keyValLogic(keyVal, c, globals.SecondLevelSplit, &possibleKeyMappings)
}

func (c *Contact) Set(value string) error {
	err := setLogic(value, c, globals.FirstLevelSplit)
	// Here we count how many empty fields we have using reflection. This is so we can default the Company value
	// To FirstName + LastName if no Company value is given.
	if err != nil && !strings.Contains(err.Error(), "cannot find key in text") || err == nil  {
		valueOf := reflect.ValueOf(*c)
		clear := true
		for i := 0; i < reflect.TypeOf(Contact{}).NumField(); i++ {
			val := reflect.Indirect(valueOf).Field(i)
			switch reflect.TypeOf(Contact{}).Field(i).Name {
			case "Address":
				if len(val.Interface().([]string)) == 0 {
					clear = false
				}
			case "Email":
				// We validate the email address given
				if !misc.IsEmailValid(val.String()) {
					return errors.New(fmt.Sprintf("\"%s\" is not a valid email", val.String()))
				}
				fallthrough
			case "Company":
				if len(val.String()) == 0 {
					c.Company = fmt.Sprintf("%s %s", c.FirstName, c.LastName)
				}
			default:
				if len(val.String()) == 0 {
					clear = false
				}
			}
		}

		// Here we clear the error only if all the required fields are set
		if clear {
			err = nil
		}
	}
	return err
}

type Bank struct {
	Bank      string
	AccountNo string
	SortCode  string
}

func (b *Bank) String() string {
	return fmt.Sprintf(`Bank details:
%s
A/c No.    %s
Sort code: %s
`, b.Bank, b.AccountNo, b.SortCode)
}

func (b *Bank) KeyVal(keyVal string) (interface{}, error) {
	possibleKeyMapping := map[interface{}]map[string]struct{} {
		&b.Bank: {
			"bank": {},
			"b": {},
		},
		&b.AccountNo: {
			"accountno": {},
			"account": {},
			"a/c no.": {},
			"a/c": {},
			"a": {},
			"no": {},
			"acc": {},
		},
		&b.SortCode: {
			"sortcode": {},
			"sort": {},
			"code": {},
			"s": {},
		},
	}
	return keyValLogic(keyVal, b, globals.SecondLevelSplit, &possibleKeyMapping)
}

// Set the Bank value From the given string value.
//
// If the given string cannot be parsed then an error will be returned otherwise the error will be nil.
//
// A valid string value can be:
//  Bank: Bank o' Clock, account: 12312312,sort: 69/69/69
func (b *Bank) Set(value string) error {
	err := setLogic(value, b, globals.FirstLevelSplit)
	// Here we validate all the fields that have been set To make sure they are good values.
	valueOf := reflect.ValueOf(*b)
	if err != nil && !strings.Contains(err.Error(), "cannot find key in text") || err == nil {
		for i := 0; i < reflect.TypeOf(Bank{}).NumField(); i++ {
			val := reflect.Indirect(valueOf).Field(i)
			switch reflect.TypeOf(Bank{}).Field(i).Name {
			case "SortCode":
				valStr := val.String()
				errPrefix := fmt.Sprintf("\"%s\" is not a valid sort code", valStr)
				var errStr string
				if len(valStr) != 6 {
					errStr += " (6 digits)"
				}
				if !str.IsNumeric(valStr) {
					errStr += " (not numeric)"
				}
				if errStr != "" {
					return errors.New(errPrefix + errStr)
				}
			case "AccountNo":
				valStr := val.String()
				errPrefix := fmt.Sprintf("\"%s\" is not a valid account number", valStr)
				var errStr string
				if len(valStr) != 8 {
					errStr = " (8 digits)"
				}
				if !str.IsNumeric(valStr) {
					errStr += " (not numeric)"
				}
				if errStr != "" {
					return errors.New(errPrefix + errStr)
				}
			default:
				break
			}
		}
	}
	return err
}

type Date time.Time

func (d *Date) String() string {
	t := time.Time(*d)
	return fmt.Sprintf("%s %s, %d",
		t.Month().String(),
		ints.Ordinal(t.Day()),
		t.Year(),
	)
}

// Set the Date value From the given string value.
//
// If the given string cannot be parsed then an error will be returned otherwise the error will be nil.
//
// A valid string value can be:
//  1/12/2000
// Note that the format is:
//  Day/Month/Year
func (d *Date) Set(value string) error {
	var err error = nil
	var t time.Time
	t, err = time.Parse("2/1/2006", value)
	*d = Date(t)
	return err
}
