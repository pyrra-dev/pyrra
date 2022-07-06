/*
Copyright 2021 Pyrra Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func cmdGenerate(logger log.Logger, configFiles, prometheusFolder string) int {
	fs, err := ioutil.ReadDir(configFiles)
	if err != nil {
		level.Error(logger).Log("msg", "failed to read config-files directory", "err", err)
		return 1
	}
	for _, file := range fs {
		if !file.IsDir() {
			_, err := objectiveAsRuleFile(filepath.Join(configFiles, file.Name()), prometheusFolder)
			if err != nil {
				level.Error(logger).Log("msg", "failed generating rule files", "err", err)
				return 1
			}
		}
	}
	return 0
}
