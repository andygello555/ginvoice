package globals

import (
	"fmt"
	"regexp"
)

const (
	FirstLevelSep  = ","
	SecondLevelSep = ";"
	ThirdLevelSep  = "|"
	KeyValueSep    = ":"
	SplitFormat    = "%s ?"
)

var (
	// FirstLevelSplit (const)
	FirstLevelSplit  = regexp.MustCompile(fmt.Sprintf(SplitFormat, FirstLevelSep))
	// SecondLevelSplit (const)
	SecondLevelSplit = regexp.MustCompile(fmt.Sprintf(SplitFormat, SecondLevelSep))
	// ThirdLevelSplit (const)
	ThirdLevelSplit  = regexp.MustCompile(fmt.Sprintf(SplitFormat, ThirdLevelSep))
	// KeyValueSplit (const)
	KeyValueSplit    = regexp.MustCompile(fmt.Sprintf(SplitFormat, KeyValueSep))
)
