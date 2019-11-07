package vltypes

import (
	"errors"
)

var (
	ErrInvalidConfigType = errors.New("vltypes: invalid type of config object")
)

// NormalizeConfig make sure config object meets basic requirement to be as map[string]interface{}
func NormalizeConfig(cfg interface{}) (map[string]interface{}, error) {
	switch r := cfg.(type) {
	case map[string]interface{}:
		return r, nil
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, v := range r {
			key, ok := k.(string)
			if !ok {
				return nil, ErrInvalidConfigType
			}

			res[key] = v
		}

		return res, nil
	}

	return nil, ErrInvalidConfigType
}
