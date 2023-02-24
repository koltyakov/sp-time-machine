# sp-time-machine

> Incrementally backup SharePoint Lists across tenants

Use it as a sample use it as a tool, use it as a base for your own project.

## Features

- Incremental backup of SharePoint Lists across tenants
- Serverless damn cheap runtime (optimized for Azure Functions)
- Precise configuration which lists and fields to sync
- Extremely [lightweight and robust changeset detection](https://github.com/koltyakov/spsync)

## Moving parts to be aware of

- SharePoint automation using Golang
- Azure Functions
- IaC using Terraform

Powered by [gosip](https://github.com/koltyakov/gosip).
