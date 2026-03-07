package folder

import "encoding/json"

// DecodeConfig converts generic map config into a typed options struct.
func DecodeConfig(cfg map[string]any, out any) error {
	if cfg == nil {
		cfg = map[string]any{}
	}
	buf, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, out)
}
