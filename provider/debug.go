package provider

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

func DebugOutput(serializable interface{}, path string) {
	var output []byte
	buf := new(bytes.Buffer)

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(serializable)

	output = buf.Bytes()
	ioutil.WriteFile(path, output, 0644)
}
