package utils

// IndexOfStringArray gets items index
func IndexOfStrArr(items []string, key string) int {
	for k, v := range items {
		if key == v {
			return k
		}
	}
	return -1 // not found
}

// IndexOfIntArray gets items index
func IndexOfIntArr(items []int, key int) int {
	for k, v := range items {
		if key == v {
			return k
		}
	}
	return -1 // not found
}
