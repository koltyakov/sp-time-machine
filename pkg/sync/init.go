package sync

import (
	"os"

	"github.com/koltyakov/sp-time-machine/pkg/config"
	"github.com/koltyakov/sp-time-machine/pkg/state"
	localState "github.com/koltyakov/sp-time-machine/pkg/state/local"
	tablesState "github.com/koltyakov/sp-time-machine/pkg/state/tables"
)

// NewState initiates sync state provider
func NewState(settings *config.Settings) (state.State, error) {
	// Switching state storage to tables for Azure Functions runtime
	if checkAzureFunctionsRuntime() {
		settings.State = "tables"
	}

	// Azure Storage Account Tables target storage
	if settings.State == "tables" {
		azureConnectionString := os.Getenv("STORAGE_STATE_CONNSTR")
		if len(azureConnectionString) == 0 {
			azureConnectionString = os.Getenv("STORAGE_ACCOUNT_CONNSTR")
		}
		// Use default Environment variable from Azure Functions environment
		if len(azureConnectionString) == 0 {
			azureConnectionString = os.Getenv("AzureWebJobsStorage")
		}
		return tablesState.NewState(azureConnectionString)
	}

	// By default use local state
	return localState.NewState()
}

// checkAzureFunctionsRuntime checks if the runtime is Azure Functions
func checkAzureFunctionsRuntime() bool {
	runtime := os.Getenv("FUNCTIONS_WORKER_RUNTIME")
	return runtime == "custom"
}
