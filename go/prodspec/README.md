# Prodspec

Prodspec is a database that encodes all the information relevant to a SCION deployment.

## File format

Prodsepc file is a TOML file that looks something like this:

```
Generator = "prodgen"
GeneratorVersion = "ana-v0.3.2-428-gb405c60-dirty"
GeneratorBuildChain = "go version go1.9.4 linux/amd64"

[Organization]
  [Organization.in7]
    Site = ["bsl1.in7", "lpf1.in7", "win1.in7"]
    AS = ["ff00:0:110"]

[AS]
  [AS."ff00:0:110"]
    Organization = "in7"
    ISD = ["64"]
```

On the top level there are few properties describing how was the layout file generated.
If you are writing a layout file by hand you are supposed to omit those.

The rest of the file has a flat structure. The node names are composed of typename (e.g. "Organization")
and instance ID (e.g. "in7").

To find out which nodes and properties exist check `go/prodspec/schema/schema.yml`.
