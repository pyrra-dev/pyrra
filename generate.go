/*
Copyright 2023 Pyrra Authors.
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
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func cmdGenerate(logger log.Logger, configFiles, prometheusFolder string, genericRules, operatorRule bool) int {
	filenames, err := filepath.Glob(configFiles)
	if err != nil {
		level.Error(logger).Log("msg", "getting file names", "err", err)
		return 1
	}

	for _, file := range filenames {
		err := writeRuleFile(logger, file, prometheusFolder, genericRules, operatorRule)
		if err != nil {
			level.Error(logger).Log("msg", "generating rule files", "err", err)
			return 1
		}
	}
	return 0
}
