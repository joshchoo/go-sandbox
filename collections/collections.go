package collections

func Map[Element any, Result any](s []Element, f func(e Element) Result) []Result {
	res := make([]Result, 0, len(s))
	for _, e := range s {
		res = append(res, f(e))
	}
	return res
}

func Filter[Element any](s []Element, f func(e Element) bool) []Element {
	res := make([]Element, 0, len(s))
	for _, e := range s {
		if f(e) {
			res = append(res, e)
		}
	}
	return res
}

func Equals[Element comparable](s []Element, other []Element) bool {
	if len(s) != len(other) {
		return false
	}
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	for i, v := range s {
		if v != other[i] {
			return false
		}
	}
	return true
}
