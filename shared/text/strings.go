package text

import "strings"

func Coalesce(first string, other ...string) string {
	res := first
	for i := range other {
		if res != "" {
			break
		}
		res = other[i]
	}
	return res
}

func SplitRemoveEmpty(in string, separator string) []string {
	var res []string
	for _, item := range strings.Split(in, separator) {
		if item != "" {
			res = append(res, item)
		}
	}
	return res
}
