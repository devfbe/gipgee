package pipelinemodel

import (
	"fmt"
	"os"
)

func (pipeline *Pipeline) WritePipelineToFile(path string) error {
	yamlString := pipeline.Render()
	fmt.Print("Generated pipeline is:\n" + yamlString)
	err := os.WriteFile(path, []byte(yamlString), 0600)
	return err
}
