package main

import (
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func cmdGenerate(logger log.Logger, configFiles, prometheusFolder string) int {
	filenames, err := filepath.Glob(configFiles)
	if err != nil {
		level.Error(logger).Log("getting files names: %w", err)
	}
	for _, f := range filenames {
		// Omit setting Objectives because UI is not used for the generate command
		_, err := processFileSLO(f, prometheusFolder)
		if err != nil {
			level.Error(logger).Log("failed to convert files names: %s: %w", f, err)
			return 2
		}
	}
	return 0
}
