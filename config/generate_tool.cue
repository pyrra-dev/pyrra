package config

import (
	"encoding/yaml"
	"tool/file"
)

command: "api.yaml": task: write: file.Create & {
	filename: "config/api.yaml"
	contents: yaml.MarshalStream([ for o in api {o}])
}

command: "kubernetes.yaml": task: write: file.Create & {
	filename: "config/kubernetes.yaml"
	contents: yaml.MarshalStream([ for o in kubernetes {o}])
}
