package commands

type ExitError struct{}

func (e *ExitError) Error() string { return "exit" }

func IsExitError(err error) bool {
	_, ok := err.(*ExitError)
	return ok
}
