# encode

Authenticate to various sources, generate a report, and save the output as CSV

## Configure

a YAML file `encode.yaml` is required for this utility. See an example YAML file at [encode.example.yaml](./encode.example.yaml).

First, define a list of `connections`. A connection is a way to authenticate to a remote service and generate a report from that service.

Next, define a list of `report`. Each report needs the following defined:

- A reference to a connection name defined in `connections`
- a cron schedule set how often the report will run
- optionally a go template file to run the report results through before saving to CSV
- TODO: validation definitons to ensure sane data is returned
