# Cloud deployment to Azure Functions environment

## Prerequisites

- [Azure Functions Core Tools, v3](https://www.npmjs.com/package/azure-functions-core-tools)
- [Go SDK](https://golang.org/dl/)
- [Visual Studio Code](https://code.visualstudio.com) & [Azure Functions extension](https://marketplace.visualstudio.com/items?itemName=ms-azuretools.vscode-azurefunctions)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) [optional]

> Mac or Linux machine, or Windows WSL2 are preferred.
> All the commands within the sample assumed running in `bash`.

## Create Azure Function Resource

> This flow is dev friendly when it comes to quick experiments. For production deployments please consider [infrastructure as code and automation](./infra) via DevOps.

In VS Code with Azure Functions extension installed:

- `CMD+Shift+P` type `Azure Functions: Create Function App in Azure... (Advanced)` and follow the wizard (further is the example of the steps)
- Select dev subscription
- Enter a globally unique name for the new function app
- Select a runtime stack: `Custom Handler`
- Select an OS: `Linux`
- Select a hosting plan: `Consumption`
- Create a new resource group
- Enter the name of the new resource group
- Create new storage account
- Enter the name of the new storage account
- Create new Application Insights resource
- Enter the name of the new Application Insights resource
- Select a location for new resources (closer to your location)

> Don't forget to remove the resource group after experiments to avoid any unpredictable charges

## Project structure highlights

- `.vscode/` - VSCode settings folder, mostly standard, however, `settings.json` has important modification `"azureFunctions.deploySubpath": "./functions"` which changes root folder for functions deployment via `ms-azuretools` extension
- `functions/` - everything related to Azure Functions, isolation makes things simpler while no mixing with custom handler related project files
  - Functions folders, containing `function.json` file
  - `.funcignore` - functions ignore file with handy ignore options
  - `host.json` - host configuration file, mostly standard, interesting and important section is `customHandler`, especially `"enableForwardingHttpRequest": true` which is not added by default
  - `local.settings[.sample].json` - contains configuration which is needed for local development and debug, local settings should be excluded from Git to avoid secrets leak
  - `proxies.json` - bypassing endpoints to external resources
- `handlers/` - Handlers sub package grouping by extensible handlers methods
- `cmd/server/main.go` - Custom Handler Go server entry point
- `cmd/server/mux.go` - routes and handler binding configuration
- `cmd/server/config.jsonc` - Workfront entitie sync configuration
- `Makefile` - useful tasks/commands collection for project automation

## Build and deploy

1\. Run build command:

```bash
make build-fns
```

The assumption is that a Linux plan was selected for the function app.

The build creates `bin/server` binary with the Go custom handler.

2\. In VSCode,

- `CMD+Shift+P` type `Azure Functions: Deploy to Function App...`
- Select subscription
- Select Function App in Azure
- Confirm deployment

## Publish via CLI

Publishing via CLI is a faster and more robust scalable way.

```bash
make appname="functionapp_name" publish
```

where `"functionapp_name"` is the name of your Function App instance.

> `az` CLI is required.

## Overriding host's settings

Azure Functions use .Net native approaches dealing with settings. Hierarchical settings can be overriden with environment variables.

E.g., you need `server.exe` locally but a Linux Functions host still needs `server` (without extention) in `hosts.json`, it can be redefined in `local.settings.json` with a value:

```jsonc
{
  // ...
  "Values": {
    // ...
    "AzureFunctionsJobHost__customHandler__description__defaultExecutablePath": "bin/server.exe"
  }
}
```

## Adding new function

1\. `CMD+Shift+P` type `Azure Functions: Create Function...`

2\. Select a template for your function, e.g. `HTTP trigger`

3\. Provide a function name. e.g. `HttpTrigger1`

4\. Select authorization level, e.g. `Function`

A function's folder and bindings definition is created `functions/HttpTrigger1/function.json`:

```json
{
  "bindings": [
    {
      "authLevel": "function",
      "type": "httpTrigger",
      "direction": "in",
      "name": "req",
      "methods": ["get", "post"]
    },
    {
      "type": "http",
      "direction": "out",
      "name": "res"
    }
  ]
}
```

5\. Run `make start-fns` and follow via `http://localhost:7071/api/HttpTrigger1` for just created function's binding

The Echo service handler will response with URL path `/api/HttpTrigger1`, that's because created function is not processed with specific route Go custom handler yet.

`Ctrl+C` to stop running functions host.

6\. Create `handlers/HttpTrigger1.go` and paste:

```golang
package handlers

import (
	"fmt"
	"net/http"
)

// HTTPTrigger1 handles a function's logic
func (h *Handlers) HTTPTrigger1(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	fmt.Fprintf(w, "Hello, %s!", name)
}
```

7\. Bind a handler with a route in `cmd/server/mux.go` in API routes section:

```golang
http.HandleFunc("/api/HttpTrigger1", h.HTTPTrigger1)
```

8\. Run `make start-fns` and navigate to `http://localhost:7071/api/HttpTrigger1?name=Function`

`Hello, Function!` should be responded back.

## Reference

- [Azure Functions custom handlers](https://docs.microsoft.com/en-us/azure/azure-functions/functions-custom-handlers)
- [Create a Go function in Azure using VSCode](https://docs.microsoft.com/en-us/azure/azure-functions/create-first-function-vs-code-other)
- [Functions Custom Handlers (Go)](https://github.com/Azure-Samples/functions-custom-handlers/tree/master/go)
- [Serverless Go in Azure Functions with custom handlers (video)](https://m.youtube.com/watch?v=RPCEH247twU)
