package lib

func MapToStringSlice(m map[string][]string) ([]string, error) {
	var result []string
	for _, value := range m {
		result = append(result, value...)
	}
	return result, nil
}
