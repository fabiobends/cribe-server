package utils

// RequiredField represents a field that needs to be validated
type RequiredField struct {
	Name  string
	Value string
}

// ValidateRequiredFields checks if any of the required fields are empty and returns their names
func ValidateRequiredFields(fields ...RequiredField) []string {
	var missingFields []string
	for _, field := range fields {
		if field.Value == "" {
			missingFields = append(missingFields, field.Name)
		}
	}
	return missingFields
}

// ValidateRequiredField is a helper function to create a RequiredField
func ValidateRequiredField(name, value string) RequiredField {
	return RequiredField{
		Name:  name,
		Value: value,
	}
}
