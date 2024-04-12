package smapper

type Option func(*Mapper)

func WithCallbacks(callbacks ...*Callback) Option {
	return func(mapper *Mapper) {
		for _, callback := range callbacks {
			mapper.callbacks[callback.Name] = callback.Func
		}
	}
}

func WithValidators(validators ...*Validator) Option {
	return func(mapper *Mapper) {
		for _, validator := range validators {
			mapper.validators[validator.Name] = validator.Func
		}
	}
}

// WithOverrideDefaultValidators if you set this option, you can override default validators (e.g. required, len, etc.)
// and your validator func will being use instead.
func WithOverrideDefaultValidators() Option {
	return func(mapper *Mapper) {
		mapper.OverrideDefaultValidators = true
	}
}

// WithIgnoreMissingValidators if you set this option, you can ignore errors that occur when you're trying to use
// a missing validator in a field's tag.
func WithIgnoreMissingValidators() Option {
	return func(mapper *Mapper) {
		mapper.IgnoreMissingValidators = true
	}
}

// WithIgnoreMissingCallbacks if you set this option, you can ignore errors that occur when you're trying to use
// a missing callback in a field's tag.
func WithIgnoreMissingCallbacks() Option {
	return func(mapper *Mapper) {
		mapper.IgnoreMissingCallbacks = true
	}
}

// WithAutoStringToNumberConversion if you set this option, you can automatically convert
// strings to numbers (int, uint and floats)
func WithAutoStringToNumberConversion() Option {
	return func(mapper *Mapper) {
		mapper.AutoStringToNumberConversion = true
	}
}

// WithAutoNumberToStringConversion if you set this option, you can automatically convert
// numbers (int, uint and floats) to strings.
func WithAutoNumberToStringConversion() Option {
	return func(mapper *Mapper) {
		mapper.AutoNumberToStringConversion = true
	}
}
