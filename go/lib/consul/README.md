# Consul

If the topology is created with the `--consul` flag, the generator
adds the appropriate consul configuration files.

The topology has one global consul server, and one consul agent 
per AS. Currently the consul server runs on `127.0.0.1` and the
per-AS agents on the public address of the first certificate server.

The agents expose the web ui on port `8500` when running. To see all currently 
running services go to [services](http://127.0.0.1:8500/ui/scion/services).

The agents add a service description for each beacon, certificate and
path server in their AS with TTL check enabled. 

The base service names are:
- `BeaconService`
- `CertificateService`
- `PathService`

In the test topology the services are prefixed with the ISD-AS identifier
to enable separation between ASes. Thus, `CertificateService` in AS 
`1-ff00:0:110` becomes `1-ff00:0:110/CertificateService`. This might also
be useful in an actual deployment where the prefix could simply be `scion`.
