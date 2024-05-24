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
	"fmt"
	"os"
	"path/filepath"

	kpt "github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func cmdKustomize(logger log.Logger, resourceList []byte, operatorRule bool) int {
	k, err := kpt.ParseResourceList([]byte(resourceList))
	if err != nil {
		level.Error(logger).Log("msg", "error parsing resource list", "err", err)
		return 1
	}

	inputDir, err := os.MkdirTemp("/tmp", "input")
	if err != nil {
		level.Error(logger).Log("msg", "creating temporary directory", "err", err)
	}
	defer os.RemoveAll(inputDir)

	outputDir, err := os.MkdirTemp("/tmp", "output")
	if err != nil {
		level.Error(logger).Log("msg", "creating temporary directory", "err", err)
	}
	defer os.RemoveAll(outputDir)

	for _, item := range k.Items {
		slo := pyrrav1alpha1.ServiceLevelObjective{}
		err = item.SubObject.As(&slo)
		if err != nil {
			level.Error(logger).Log("msg", "failed to marshal resource", "err", err)
		}

		bytes, err := yaml.Marshal(slo)
		if err != nil {
			level.Error(logger).Log("msg", "failed to marshal resource", "err", err)
			return 1
		}

		inputFile := fmt.Sprintf("%s/%s.yaml", inputDir, item.GetId().Name)
		// level.Info(logger).Log("msg", "creating temporary file", "file", inputFile)

		if err := os.WriteFile(inputFile, bytes, 0o644); err != nil {
			level.Error(logger).Log("msg", "failed to write temp input file", "err", err)
			return 1
		}
	}

	filenames, err := filepath.Glob(fmt.Sprintf("%s/*.yaml", inputDir))
	if err != nil {
		level.Error(logger).Log("msg", "getting file names", "err", err)
		return 1
	}

	genericRules := k.FunctionConfig.SubObject.GetBool("genericRules")
	for _, file := range filenames {
		err := writeRuleFile(logger, file, outputDir, genericRules, operatorRule)
		if err != nil {
			level.Error(logger).Log("msg", "generating rule files", "err", err)
			return 1
		}
	}

	filenames, err = filepath.Glob(fmt.Sprintf("%s/*.yaml", outputDir))
	if err != nil {
		level.Error(logger).Log("msg", "getting file names", "err", err)
		return 1
	}

	output := kpt.ResourceList{}
	output.FunctionConfig = kpt.NewEmptyKubeObject()

	for _, file := range filenames {
		o, err := os.ReadFile(file)
		if err != nil {
			level.Error(logger).Log("msg", "reading output file", "file", "file", "err", err)
			return 1
		}

		rule, _ := kpt.ParseKubeObject(o)
		output.UpsertObjectToItems(rule, nil, true)
	}

	bytes, err := output.ToYAML()
	if err != nil {
		level.Error(logger).Log("msg", "failed to marshal resource", "err", err)
		return 1
	}
	fmt.Print(string(bytes))

	return 0
}
