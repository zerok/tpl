package world

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Data can be used to store arbitrary data (e.g. coming from data-files).
type Data map[string]interface{}

// LoadData fills a newly created Data object based on the given definitions.
func LoadData(datadefs []string, cwd string) (Data, error) {
	result := Data{}
	for _, datadef := range datadefs {
		elems := strings.SplitN(datadef, "=", 2)
		if len(elems) != 2 {
			return nil, errors.Errorf("invalid data definition `%s`", datadef)
		}
		key := elems[0]
		file := elems[1]
		var value interface{}
		fp, err := os.Open(filepath.Join(cwd, file))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file in `%s`", datadef)
		}
		switch filepath.Ext(file) {
		case ".yaml":
			err = loadYAMLValue(&value, fp)
		case ".yml":
			err = loadYAMLValue(&value, fp)
		case ".json":
			err = json.NewDecoder(fp).Decode(&value)
		default:
			return nil, errors.Errorf("unsupported file-extension in `%s`", datadef)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "data parsing failed for `%s`", datadef)
		}
		result[key] = value
	}
	return result, nil
}

func loadYAMLValue(out *interface{}, fp io.Reader) error {
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}
