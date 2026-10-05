package ports

import "encoding/json"

// Optional retains absent, null, and populated API fields. The false
// key represents null; an empty map is omitted by encoding/json.
type Optional[T any] map[bool]T

func (v *Optional[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*v = Optional[T]{false: *new(T)}
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = Optional[T]{true: value}
	return nil
}
func (v Optional[T]) MarshalJSON() ([]byte, error) {
	if value, ok := v[true]; ok {
		return json.Marshal(value)
	}
	return []byte("null"), nil
}
