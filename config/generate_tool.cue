package config

import (
	"encoding/yaml"
	"tool/file"
)

command: "api.yaml": task: write: file.Create & {
	filename: "config/api.yaml"
	contents: yaml.MarshalStream([ for o in api {o}])
}

command: "manager.yaml": task: write: file.Create & {
	filename: "config/manager.yaml"
	contents: yaml.MarshalStream([ for o in manager {o}])
}
