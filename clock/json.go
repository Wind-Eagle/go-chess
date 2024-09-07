package clock

import (
	"encoding/json"
	"fmt"
)

func (c Control) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Control) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var cs string
	if err := json.Unmarshal(data, &cs); err != nil {
		return err
	}
	res, err := ControlFromString(cs)
	if err != nil {
		return fmt.Errorf("time control from string: %w", err)
	}
	*c = res
	return nil
}
