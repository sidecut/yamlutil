package yamlutilities

import (
	"errors"
	"io"

	"github.com/sidecut/yamlutil/config"
	"gopkg.in/yaml.v2"
)

func GetYamlMap(input io.Reader) (GenericMap, error) {
	decoder := yaml.NewDecoder(input)

	// when i = 0, store off the value
	// when i = 1 and we get an EOF, return the value
	// otherwise it's an error
	var yamlmaps []GenericMap
	for i := 0; ; i++ {
		var yamlMap GenericMap
		err := decoder.Decode(&yamlMap)
		if i == 0 {
			if err != nil {
				// decoding error of some sort
				return nil, err
			}

			// Store off value for use later
			yamlmaps = append(yamlmaps, yamlMap)
		} else {
			// Not our first rodeo
			if err == io.EOF || config.Strict == false {
				// Cool! Only one result found, so return it
				return yamlmaps[0], nil
			}

			// NOT eof, so this is wrong
			return nil, errors.New("Multiple documents encountered")
		}
	}
}
