# Logging

## Azure Functions

With Azure Functions, use Application Insights logging query:

```kql
traces
| where message startswith "{"
| extend d=parse_json(message)
| project d.timestamp, d.level, d.message
```

## On-Premise

In On-Premise running mode, logs are written to `./logs` folder.