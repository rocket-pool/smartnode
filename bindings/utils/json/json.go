package json

import (
	"encoding/json"
	"fmt"
)

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("%w\nUnable to Unmarshal JSON string %s", err, string(data))
	}

	return nil
}
