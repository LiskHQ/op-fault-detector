// CLI to run fault detector service
package main

import (
	"github.com/LiskHQ/op-fault-detector/pkg/log"
)

func main() {
	logger, err := log.NewDefaultProductionLogger()

	if err != nil {
		panic(err)
	}

	// To be removed after implementation
	logger.Info("Running fault detector")
}
