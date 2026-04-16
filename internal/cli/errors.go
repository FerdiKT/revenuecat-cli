package cli

type CLIError struct {
	Code    int
	Message string
}

func (e *CLIError) Error() string {
	return e.Message
}
