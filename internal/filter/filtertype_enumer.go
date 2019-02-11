// Code generated by "enumer -type=FilterType"; DO NOT EDIT.

package filter

import (
	"fmt"
)

const _FilterTypeName = "LintTidy"

var _FilterTypeIndex = [...]uint8{0, 4, 8}

func (i FilterType) String() string {
	if i < 0 || i >= FilterType(len(_FilterTypeIndex)-1) {
		return fmt.Sprintf("FilterType(%d)", i)
	}
	return _FilterTypeName[_FilterTypeIndex[i]:_FilterTypeIndex[i+1]]
}

var _FilterTypeValues = []FilterType{0, 1}

var _FilterTypeNameToValueMap = map[string]FilterType{
	_FilterTypeName[0:4]: 0,
	_FilterTypeName[4:8]: 1,
}

// FilterTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func FilterTypeString(s string) (FilterType, error) {
	if val, ok := _FilterTypeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to FilterType values", s)
}

// FilterTypeValues returns all values of the enum
func FilterTypeValues() []FilterType {
	return _FilterTypeValues
}

// IsAFilterType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i FilterType) IsAFilterType() bool {
	for _, v := range _FilterTypeValues {
		if i == v {
			return true
		}
	}
	return false
}