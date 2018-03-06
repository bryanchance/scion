## SIG Config UI

### Sites configuration page
**Name**. Unique identifier for the site.

**VHost**. Virtual host name or IP address. This will be the address of the host
in a simple setup, or a virtual hostname or IP in a high-availability setup.

**Hosts**. Comma separated list of machine hostnames or IPs. In a simple setup,
this will be identical to VHost. In a high-availability setup, each machine
should be included here.

### ASes configuration page

This page can be used to configure a single site. It includes global
information about the site at the top, and an AS configuration panel for each
defined AS.

**Management**. Clicking the `Push and reload config` button will generate a
new JSON SIG Config file from saved settings, copy it to each host in the site
and then send a reload signal to the running SIG. Finally, the SIG is queried
via HTTP to verify that it has loaded the correct config.

**Remote ASes**. The list of ASes the SIG in the current site is aware of.

**Path selectors**. Path selectors can be used to choose a subset of paths from
a larger set. For example, if a SIG can reach a remote AS via two paths, one
through ISD-AS 2-80 and one through ISD-AS 2-81, a path selector can be used to
tell the SIG to always prefer the one through AS80.  A filter is a comma
separated string of ISD-AS#IFID tokens, each token representing a condition.
Wildcards can be specified for via a 0 value for an ISD, AS or IFID. Some
examples:

* To select paths going through ISD-AS 2-81: `2-81#0`;
* To select paths going through ISD 5 and then ISD 6 (not necessarily one after
  the other): `5-0#0,6-0#0`;

Ordering of tokens is important, but gaps are allowed:

* `5-1#10,6-2#20` does not match a path containing `...,6-2#20,5-1#10,...`;
* `5-1#10,6-2#20` does match `...,5-1#10,7-1#52,7-1#55,6-2#20,...`;


#### Remote AS configuration

**Remote IPv4 networks**. The list of IPv4 prefixes known to belong to the
remote AS. These will be converted to routes to the remote AS by the SIG.

**Remote SIGs**. The list of available SIGs for the remote AS. Multiple SIGs
can be defined to increase reliability. If one remote SIG is down, the local
SIG will start forwarding traffic to the remaining available ones. The
**encapsulation port** and **control port** are used by the SIGs to forward
traffic and detect availability, and should be reachable by the local SIG.

**Sessions**. Sessions tell the SIG which paths to use for which types of
traffic. Each session represents a packet queue. Classification rules
(explained below) map packets to sessions, and each session selects paths
according to a predefined **Path selector** (explained above).

**Packet classifier rules**. Contains traffic policy rules, one per line. One
type of rule is supported, which tells packets matching certain criteria which
SCION path to use. All fields need to be satisfied for the rule to take effect:

```
  name=NAME [ src=NETWORK ]  [ dst=NETWORK ] [ dscp=DSCP ] sessions=SESSIONS
```

`NAME` is used to link the class of traffic to the corresponding policy. Names
can only contain letters, digits and underscores.

`NETWORK` must be specified in CIDR format, e.g., 172.16.0.0/24.

`DSCP` can be specified in decimal, octal (via a leading 0) or hexadecimal (via
a leading 0x) notation.

`SESSIONS` is a comma-separated list of SCION Session identifiers in descending
priority, e.g., `1,2,3`. If a session is not available (i.e., due to no paths
matching its path selector), the next session is used.

For example, the following command tells traffic coming from network
172.16.0.0/24 and going to network 192.168.0.0/16, with DSCP bits set to 0x12
to use sessions 1, 2, 3, and 0:
```
name=audio src=172.16.0.0/24 dst=192.168.0.0/16 dscp=0x12 sessions=1,2,3,0
```

