package smapper

type Config struct {
	// it allows you to override the default validators (e.g. eq, required, etc.) and use your
	// custom validator. if it's false (by default), then it ignores your validator and executes
	// the default validator.
	OverrideDefaultValidators bool
	// if you specify a validator in a field's tag that does not exist, you will get an error (by default),
	// but this allows you to ignore those missing validators.
	IgnoreMissingValidators bool
	// if you specify a callback in a field's tag that does not exist, you will get an error (by default),
	// but this allows you to ignore those missing callbacks.
	IgnoreMissingCallbacks bool
	// if you try to map a string value to a numeric value (int, uint or float), you will get an error (by default),
	// but this allows you to automatically convert strings to numbers.
	AutoStringToNumberConversion bool
	// if you try to map a numeric value (int, uint or float) to a string, you will get an error (by default),
	// but this allows you to automatically convert numbers to strings.
	AutoNumberToStringConversion bool
}
