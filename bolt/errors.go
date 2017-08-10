package bolt

import (
	"fmt"
)

// FactoidAlreadyExistsError Represents inserting an already existing factoid response.
type FactoidAlreadyExistsError string

func (e FactoidAlreadyExistsError) Error() string {
	return fmt.Sprint("FactoidAlreadyExistsError: \"", e, "\"")
}
