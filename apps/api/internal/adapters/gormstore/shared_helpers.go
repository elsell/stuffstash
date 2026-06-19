package gormstore

func stringPtr(value string) *string {
	return &value
}

func stringFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
