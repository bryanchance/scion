## Running

Make sure the config output directory `output` exists in the running folder. And you have a config file.

Then run the interface e.g. with `go run main.go -id=sm4-21-9 -bind=:8080 -config=cfg.json`.

## Configuration

The SIG Policy Configurator needs a json config file to run. The following keys are defined:
* **Features** (integer). The feature set the web UI should expose to users. Two
  feature sets are defined:
  * **Level 0**. Sites cannot be added or deleted. Remote ASes cannot be added
    or deleted. Sessions and path filters cannot be added or deleted, and are
    not shown in the UI. Predefined Session Aliases are displayed instead. These
    aliases contain hidden sequences of sessions, and can be referenced in
    policies. For deployments using **level 0**, the database should come preloaded
    with Sites, sessions, path filters and session aliases. Note that functionality
    is only hidden; users can still run any query via crafting GETs and POSTs, but
    this is unsupported behavior.
  * **Level 1**. All functionality is exposed. Because Session Aliases are no
    longer needed, they are not displayed. For deployments using **level 1**,
    preloading the database with information is optional.
* **OutputDir**. The folder where the generated config files will be stored.
* **DBPath**. The SQlite3 database to use. Folder output must already exist.
  The database will be created if it does not already exist.
* **WebAssetRoot**. The folder where the static resources of the web app are
  located.
* **SIGCfgPath**. Path to SIG config files on target machines.

Example file:
```
{
    "Features": 1,
    "OutputDir": "output/",
    "DBPath": "testdb",
    "SIGCfgPath": "/etc/scion/sig/sig.json",
    "WebAssetRoot": "static/"
}
```
