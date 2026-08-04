package user

func optionalString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
