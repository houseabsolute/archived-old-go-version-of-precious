package filter

import "github.com/pkg/errors"

//go:generate enumer -type=FilterType
type FilterType int

const (
	Lint FilterType = iota
	Tidy
)

// There's a circular issues with this method. The enumer code generates
// FilterTypeString but if you run enumer against this code before anything
// has been generated then the parser complains that FilterTypeString is an
// undeclared name.

// UnmarshalTOML implements the toml.UnmarshalerRec interface for FilterType
func (i *FilterType) UnmarshalTOML(decode func(interface{}) error) error {
	var s string
	if err := decode(&s); err != nil {
		return errors.Wrap(err, "Error decoding FilterType value")
	}

	var err error
	*i, err = FilterTypeString(s)
	return err
}
