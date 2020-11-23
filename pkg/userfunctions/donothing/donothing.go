package donothing

import "errors"

var name = "do_nothing"

var userFunctions = map[string]interface{}{
	// sample function to prove that function registration works.
	"do_nothing": func(a ...string) (string, error) {
		if len(a) == 0 {
			return "", errors.New("do_nothing: no arg specified")
		}
		return a[0], nil
	},
}

// Name - Return the Name of the UDFs
func Name() string {
	return name
}

// Load - Return the user-defined functions
func Load() map[string]interface{} {
	return userFunctions
}