## Running

Make sure you have a config file. You should create your own TLS keys:
`openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.pem -out cert.pem`

Then run the interface e.g. with `go run main.go -id=sm4-ff00:0:2f-9 -bind=:8080 -config=cfg.json`.
Open the WebUI on `https://<your-bind-arg>`

## Configuration

The SIG Policy Configurator needs a json config file to run. The following keys are defined:

*   **Features** (integer). The feature set the web UI should expose to users. Two feature sets are
    defined:
    *   **Level 0**. Sites cannot be added or deleted. Remote ASes cannot be added or deleted. Path
        filters cannot be added or deleted, and are not shown in the UI. For deployments using
        **level 0**, the database should come preloaded with Sites and path filters. Note that
        functionality is only hidden; users can still run any query via crafting GETs and POSTs, but
        this is unsupported behavior.
    *   **Level 1**. All functionality is exposed. For deployments using **level 1**, preloading the
        database with information is optional.
*   **Key**. Secret key for WebUI Authentication
*   **Username**. Username for the WebUI
*   **Password**. Password for the WebUI
*   **TLSCertificate**. TLS Certificate can be self-signed
*   **TLSKey**. TLS Key
*   **DBPath**. The SQlite3 database to use. Folder output must already exist. The database will be
    created if it does not already exist.
*   **WebAssetRoot**. The folder where the static resources of the web app are located.
*   **SIGCfgPath**. Path to SIG config files on target machines.

Example file:

```
{
    "Features": 1,
    "Key": "secret key",
    "Username": "admin",
    "Password": "password",
    "TLSCertificate": "go/sigmgmt/cert.pem",
    "TLSKey": "go/sigmgmt/key.pem",
    "DBPath": "testdb",
    "SIGCfgPath": "/etc/scion/sig/sig.json",
    "WebAssetRoot": "static/"
}
```
