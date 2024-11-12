package testhelpers

func CheckSlicesEquality(a []any, b []any) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[any]int)
	bMap := make(map[any]int)

	for _, v := range a {
		aMap[v]++
	}
	for _, v := range b {
		bMap[v]++
	}

	for k, v := range aMap {
		if bMap[k] != v {
			return false
		}
	}

	return true
}

func StringSliceToAnySlice(s []string) []any {
	var result []any
	for _, v := range s {
		result = append(result, v)
	}
	return result
}
