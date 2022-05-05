package tables

func chunkStringSlice(slice []string, chunkSize int) [][]string {
	chunked := [][]string{}
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunked = append(chunked, slice[i:end])
	}
	return chunked
}
