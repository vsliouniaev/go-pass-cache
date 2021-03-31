package util

import "fmt"

type ArrayFlags []string

func (af *ArrayFlags) String() string {
	return fmt.Sprintf("%s", []string(*af))
}

func (af *ArrayFlags) Set(value string) error {
	*af = append(*af, value)
	return nil
}
