# SIG Configuration

This document describes how a SIG (SCION IP Gateway) can be manually configured via its JSON
configuration file.

SIGs connect IPv4 and/or IPv6 islands across SCION networks. In the most common scenario, SIGs sit
at the edge of both IP networks and the SCION network, tunneling all IP traffic by encapsulating it
in SCION packets. The health of SCION paths (that is, the sequences of SCION Autonomous Systems and
Interface IDs that packets transit) is continuously monitored, allowing SIGs to ensure fast failover
whenever network conditions degrade.

## Table of contents

*   [Basic path selection](#basic-path-selection)
    *   [Use case - Single path traffic](#use-case-1)
*   [Policy-based path selection](#policy-based-path-selection)
    *   [Traffic policies](#traffic-policies)
    *   [Traffic classes](#traffic-classes)
    *   [Sessions](#sessions)
    *   [Path selectors](#path-selectors)
    *   [Use case - Multiple paths with failover](#multiple-paths-with-failover)
        *   [Choosing different paths for different traffic](#different-paths)
        *   [Session of last resort](#session-of-last-resort)

## Basic path selection <a id="basic-path-selection"> </a>

SIGs connect IP islands from different SCION ASes. Each SIG has a list of remote ASes statically
configured by the administrator. The list of remote ASes is configured via the JSON keys in the
top-level "ASes" JSON object. For example, to define two remote ASes `2-ff00:0:2` and `2-ff00:0:3`,
use the following:

```json
{
  "ASes": {
    "2-ff00:0:2": {
      ... remote AS config omitted ...
    },
    "3-ff00:0:3": {
      ... remote AS config omitted ...
    }
  }
}
```

Note that for a SIG to communicate with a remote SIG, it needs a path to the remote SIG's AS. Paths
are automatically obtained by the local SIG by querying the local SCION Daemon.

SIGs forward IP traffic according to their internal IP routing table. The routing table is
statically configured by the network administrator (dynamic modes of operation are planned for
future versions) via name "Nets" inside each remote AS definition. Note that in JSON, the IP
addresses are surrounded by quotes. The following example states that networks `192.168.2.0/24` and
`192.168.3.0/24` exist in remote AS `2-ff00:0:2`, and that the local SIG should route traffic for
these destinations to that remote AS:

```json
{
  "ASes": {
    "2-ff00:0:2": {
      "Nets": ["192.168.2.0/24", "192.168.3.0/24"],
      ... rest of config omitted ...
    }
  }
}
```

The network prefixes must be unique, and cannot overlap with the prefixes of any other remote AS
(e.g., in this case it is unsupported for a different remote AS to have an allocation for
`192.168.0.0/16`, since it overlaps with `192.168.2.0/24` and `192.168.3.0/24`, or to have an
allocation for `192.168.2.192/26` since it overlaps with `192.168.2.0/24`).

To send traffic, the local SIG also needs to know the address of the remote endpoint of the tunnel
(that is, at least one SIG in the remote AS). This is specified via name "Sigs":

```json
{
  "ASes": {
    "2-ff00:0:2": {
      "Nets": ["192.168.2.0/24", "192.168.3.0/24"],
      "Sigs": {"Madrid": {"Addr": "192.168.2.1"}},
      ... rest of config omitted ...
    }
  }
}
```

Once the configuration is loaded, the SIG will encapsulate IP traffic for networks `192.168.2.0/24`
and `192.168.3.0/24` in SCION packets and send them to AS `2-ff00:0:2` on address `192.168.2.1`. As
opposed to other tunneling protocols, it is mandatory (and does not cause a routing loop) in SIGs
for the address of the remote SIG to be included in the network prefixes of the remote AS. This is
because SCION is used to carry all the traffic, including session establishment and health
monitoring.

### Use case #1 - Single path traffic <a id="use-case-1"> </a>

In the topology below, two SIGs from two different ASes communicate across a single SCION path. The
Berlin AS (`1-ff00:0:11`) contains one IPv4 network (`192.168.1.0/24`) that is directly connected to
a SIG. The hosts in this network (e.g., HostA) want to talk to the hosts in the London AS
(`2-ff00:0:2`), located in networks `192.168.2.0/24` and `192.168.3.0/24`. The two ASes communicate
via the path that goes through the Zürich and Paris ASes.

![Use case 1](case1.png)

Configuring a SIG is done by editing the JSON file. This can be done either via the SIG WebUI, or by
manually creating the file.

At a minimum, the config file must specify the following:

*   The known remote ASes;
*   The IP networks contained in each remote AS;
*   The list of SIGs in each remote AS.

For example, a minimal configuration file for the Berlin SIG looks like:

```json
{
    "ASes": {
        "2-ff00:0:2": {
            "Nets": ["192.168.2.0/24", "192.168.3.0/24"],
            "Sigs": { "London": { "Addr": "192.168.2.1" } }
        }
    }
}
```

The London SIG configuration file is similar:

```json
{
    "ASes": {
        "1-ff00:0:11": {
            "Nets": ["192.168.1.0/24"],
            "Sigs": { "Berlin": { "Addr": "192.168.1.1" } }
        }
    }
}
```

Once the file is in place (typically, this is `/etc/scion/sig.json`), the configuration file can be
reloaded by sending a SIGHUP signal to the SIG process.This will make the new configuration
immediately take effect. The signal can be sent manually (e.g., via `kill`) or via the following:

```bash
scion@berlin$ sudo systemctl kill -sSIGHUP sig.service
```

If the configuration could not be loaded (e.g., due to a syntax error) the SIG configuration remains
unchanged. In both cases, check the log to verify that the SIG reloaded the config successfully:

```bash
scion@berlin$ tail /var/log/scion/sig*.log
2018-04-30 15:10:35.011305+0000 [INFO] reloadOnSIGHUP: reloading...
2018-04-30 15:10:35.012406+0000 [INFO] ReloadConfig: Adding AS... ia=2-ff00:0:2
2018-04-30 15:10:35.012475+0000 [INFO] Updated SIG ia=2-ff00:0:2 sig=2-ff00:0:2,[192.168.2.1]:10081:10080
2018-04-30 15:10:35.012620+0000 [INFO] ReloadConfig: Added AS ia=2-ff00:0:2
2018-04-30 15:10:35.012658+0000 [INFO] reloadOnSIGHUP: reload done success=true
```

## Policy-based path selection <a id="policy-based-path-selection"> </a>

SIGs can be configured to forward packets across certain paths, depending on the values of various
packet fields. To implement this, SIGs define and support the following:

*   [Traffic policies](#traffic-policies)
*   [Traffic classes](#traffic-classes)
*   [Sessions](#sessions)
*   [Path selectors](#path-selectors)

Traffic flow is controlled via **traffic policies**. Each policy is composed of two elements:

*   A class name, used to identify the **traffic class** used to match the packets.
*   A list of **sessions**, in order of priority, used to tell the SIG which session should be used
    to send the packet.

A **session** is the state related to a tunneling (i.e., encapsulation) session. The session
contains information such as usable paths, the health of the active path and sequence numbers.

Each session is defined by a name (a number between 0 and 255), and a **path selector** that chooses
the paths that can be used (out of the pool of all known paths).

The following sections take a more in-depth look at these concepts, and how they can be configured.

### Traffic classes <a id="traffic-classes"> </a>

A packet is said to match a traffic class if the evaluation of the conditions in the traffic class
returns true. The conditions are evaluated on the raw content of each packet. Conditions can be
composed, thus creating arbitrarily complex classifiers.

Classes can be written using the following conditions:

*   **CondAllOf**: contains a list of subconditions. Returns true if all subconditions yield true.
    If empty, it returns true.
*   **CondAnyOf**: contains a list of subconditions and returns true if at least one subcondition
    yields true. If empty, it return true.
*   **CondNot**: contains a single subcondition; CondNot returns the negation of the result of the
    subcondition.
*   **CondBool**: can be true or false, and always evaluates to its chosen value. This is useful
    during testing.
*   **CondIPv4**: contains a condition for an IPv4 packet. Possible options are:
    *   **MatchSource**: checks whether the source address of the packet is within the network in
        field "Net".
    *   **MatchDestination**: checks whether the destination address of the packet is within the
        network in field "Net".
    *   **MatchDSCP**: checks whether the ToS/DSCP fields of the packet exactly match the value in
        field "DSCP".

Classes are defined in the top level of the JSON config file, under name "Classes".

A simple traffic class that selects traffic coming from network `192.168.1.0/24` is defined as
follows:

```json
"Classes": {
  "example-source": {
    "CondIPv4": {"MatchSource": {"Net": "192.168.1.0/24"}}
  }
}
```

Note that we also needed to specify a name for the class, in this case `example-source`. The name is
later used in traffic policies to refer to this specific class.

To select traffic coming from multiple networks, compose multiple conditions using **CondAnyOf**:

```json
"Classes": {
  "example-multiple-source": {
    "CondAnyOf": [
      {"CondIPv4": {"MatchSource": {"Net": "192.168.1.0/24"}}},
      {"CondIPv4": {"MatchSource": {"Net": "192.168.2.0/24"}}},
      {"CondIPv4": {"MatchSource": {"Net": "192.168.3.0/24"}}}
    ]
  }
}
```

**CondAnyOf** and **CondAllOf** can be used together to create more complex classes. For example, to
select either traffic going from `192.168.1.0` to `192.168.2.0`, or from `192.168.3.0` to
`192.168.4.0`:

```json
"Classes": {
  "example-src-dst-match": {
    "CondAnyOf": [
      {
        "CondAllOf": [
          {"CondIPv4": {"MatchSource": {"Net": "192.168.1.0/24"}}},
          {"CondIPv4": {"MatchDestination": {"Net": "192.168.2.0/24"}}}
        ]
      },
      {
        "CondAllOf": [
          {"CondIPv4": {"MatchSource": {"Net": "192.168.3.0/24"}}},
          {"CondIPv4": {"MatchDestination": {"Net": "192.168.4.0/24"}}}
        ]
      }
    ]
  }
}
```

### Path selectors <a id="path-selectors"></a>

A path selector filters out sets of paths depending on a set of predicates. They are similar to a
route map, except they operate on paths. Currently, the SIG accepts selectors with a single
predicate, `PP`, that specifies what ISDs, ASes and IFIDs should be contained by a path in order to
be selected.

![Selectors](case2.png)

Refer to the topology above. It similar to the topology in the previous use case, except a new
transit AS, Frankfurt, is available between Berlin and London. As opposed to the other ASes, the
Frankfurt AS has two border routers. Suppose Berlin wants to communicate with London. In this case,
the following SCION paths are available (note that depending on how Frankfurt operates internally,
more paths in addition to those listed below might exist):

```basic
1-ff00:0:11#3, 1-ff00:0:12#1, 1-ff00:0:12#2, 2-ff00:0:2#3
1-ff00:0:11#2, 1-ff00:0:12#3, 1-ff00:0:12#4, 2-ff00:0:2#2
1-ff00:0:11#1, 3-ff00:0:3#1, 3-ff00:0:3#2, 4-ff00:0:4#1, 4-ff00:0:4#2, 2-ff00:0:2#1
```

The path elements above are in ISD-AS#IFID notation. The same notation is used when defining path
selectors.

A selector specifies a certain ISD, AS and IFID that the path must contain. For example, to select
only the path going through border router Frankfurt-1 in the Frankfurt AS, we can define the
following selector:

```basic
1-ff00:0:12#1
```

This means that paths going through ISD 1, AS `ff00:0:12` and IFID 1 are selected. Paths going
through border router Frankfurt-2 are avoided, so sessions using this path selector will never
forward traffic through Frankfurt-2. If we don't care about the IFID and want to select all paths
going through any border router in the Frankfurt AS, we can replace the IFID in the path selector
with a wildcard 0:

```basic
1-ff00:0:12#0
```

Wildcards can also be used for ASes or IFIDs. For example, the match all path selector would look
like (note the shorthand AS notation for AS 0):

```basic
0-0#0
```

Selectors can include multiple comma-separated items. All the items must match parts of the path.
Gaps are allowed, but the matching must be done in order. Specifically, the Berlin to London path
that goes through Paris is matched by the following rules:

```basic
3-ff00:0:3#0,4-ff00:0:4#0
3-ff00:0:3#1,2-ff00:0:2#1
```

but is not matched by:

```basic
4-ff00:0:4#0,3-ff00:0:3#0
```

because the items are in the wrong order.

To write a selector in JSON format, the predicate string is embedded in a **CondPathPredicate**
object, with the predicate string itself added under name `PP`:

```json
"CondPathPredicate": {"PP": "4-ff00:0:4#0"}
```

Path selectors are specified in the SIG config file under the top level name "Actions". The format
is the following:

```json
"Actions": {
  "custom-name": {
    "ActionFilterPaths": {
      ... filter omitted ...
    }
  }
}
```

Only **ActionFilterPaths** is supported for now, but other actions might be added in the future. An
example with one of the predicates defined above would look like:

```json
"Actions": {
  "go-through-paris": {
    "ActionFilterPaths": {
      "CondPathPredicate": {"PP": "4-ff00:0:4#0"}
    }
  }
}
```

To create more complex predicates, the conditions can be combined the same way traffic class
conditions are. For example, to create a path selector that chooses paths that go through either ISD
1 or ISD 2:

```json
"Actions": {
  "go-through-either": {
    "ActionFilterPaths": {
      "CondAnyOf": [
        {"CondPathPredicate": {"PP": "1-0#0"}},
        {"CondPathPredicate": {"PP": "2-0#0"}}
      ]
    }
  }
}
```

One useful configuration pattern is the path avoidance pattern, which can be used to prevent routing
through certain ISDs and ASes. Suppose ISD 10 is now connected to the SCION network, but this ISD
has a history of data theft and other security issues. We want to avoid routing through this ISD.
The simplest way to do this is to only select paths that do not go through ISD 10 via a `CondNot`
condition:

```json
"Actions": {
  "avoid-isd": {
    "ActionFilterPaths": {
      "CondNot": {"CondPathPredicate": {"PP": "10-0#0"}}
    }
  }
}
```

### Sessions <a id="sessions"> </a>

Sessions describe independent SIG to SIG tunneling states. Sessions are defined per remote AS, under
name "Sessions":

```json
{
  "ASes": {
    "2-ff00:0:2": {
       "Nets": ["192.168.2.0/24", "192.168.3.0/24"],
       "Sigs": {
         "LondonA": { "Addr": "192.168.2.1" }
       },
       "Sessions": {
         ... define sessions here ...
       }
    }
  }
}
```

To select eligible paths, each session either needs to reference an **ActionFilterPaths** object, or
the empty quotes "" (meaning no filter, and all paths are used). For example, to define a session
that allows all paths and a session that uses the `go-trough-either` paths defined in the previous
section, use the following:

```json
"Sessions": {
  "0": "",
  "10": "go-through-either"
}
```

In the above example, two session are defined. The session with ID 0 will use all available paths to
the destination AS. The session with ID 10 will only use paths that go through ISD 1 or ISD 2 (as
defined in the previous section). Session IDs are only used to identify the session, and do not
represent any form of priority.

### Traffic policies <a id="traffic-policies"> </a>

Traffic policies map traffic classes to sessions. Traffic policies are configured under name
"PktPolicies" in the remote AS configuration:

```json
{
  "ASes": {
    "2-ff00:0:2": {
      "Nets": ...omitted...,
      "Sigs": ...omitted...,
      "Sessions": ...omitted...,
      "PktPolicies": [
        {
          "Classname": <classname>,
          "SessIDs": <list-of-sessions>
        }
      ]
    }
  }
}
```

In an actual configuration file, value `<classname>` above is replaced with a previously defined
traffic class (see section [Traffic classes](traffic-classes)), while value `<list-of-sessions>` is
replaced with a list of session names in decreasing order of priority. If the first session in the
list is healthy (i.e., has at least one usable path and traffic can flow across the path), it will
always be used. If the first session is unhealthy, the next one is used and so on and so forth.

To see a `PktPolicies` example, recall that in the previous sections we defined the following
sessions, path selectors and traffic classes:

```json
"Sessions": {
  "0": "",
  "10": "go-through-either"
}
```

```json
"Actions": {
  "go-through-either": {
    "ActionFilterPaths": {
      "CondAnyOf": [
        {"CondPathPredicate": {"PP": "1-0#0"}},
        {"CondPathPredicate": {"PP": "2-0#0"}}
      ]
    }
  }
}
```

```json
"Classes": {
  "example-source": {
    "CondIPv4": {"MatchSource": {"Net": "192.168.1.0/24"}}
  }
}
```

For simplicity, we do not include the whole file, but note that `Classes`, `Actions` and `Sessions`
need to be inserted at specific points in the JSON file. One example of `PktPolicies`, based on the
classes and sessions above could look like:

```json
"PktPolicies": [
  {
    "Classname": "example-source",
    "SessIDs": [10, 0],
  }
]
```

In this example, we state that IP traffic matching traffic class `example-source` should be
forwarded across session 10, if possible. Also, session 10 uses the `go-through-either` path
selector, which means that it will only allow paths going through ISD 1 (`1-0#0`) or ISD 2
(`2-0#0`). If session 10 is unhealthy, session 0 will be used. This session doesn't specify any path
selector, so any path can be used. Note that the list of session names does not include quotes.

### Use case #2 - Multiple paths with failover <a id="multiple-paths-with-failover"> </a>

This use case presents how traffic classes, path selectors, sessions and traffic policies come
together to implement path failover policies in the SIG.

SIGs monitor paths to detect network failures. As soon as a path is down, the SIG can transparently
change to the remaining healthy paths. The topology below (which is the same as the topology in the
Path selectors section) presents such a scenario.

The Berlin AS wants to communicate with the London AS. If all the paths are healthy, the path
through Zürich should be preferred. If the path through Zürich becomes unavailable, traffic should
start flowing on the paths through Frankfurt.

![Use case 2](case2.png)

To implement this, first the packets must be classified. Define one class matching any packet:

```json
"Classes": {
  "all": {
    "CondBool": true
  }
}
```

Then, define path selectors for the paths through Zürich:

```json
"go-through-zuerich": {
  "ActionFilterPaths": {
    "CondPathPredicate": {"PP": "3-ff00:0:3#0"}
  }
}
```

And a path selector for the path through Frankfurt:

```json
"go-through-frankfurt": {
  "ActionFilterPaths": {
    "CondPathPredicate": {"PP": "1-ff00:0:12#0"}
  }
}
```

Putting it all together, the "Actions" section of the configuration file should now look like:

```json
"Actions": {
  "go-through-zuerich": {
    "ActionFilterPaths": {
      "CondPathPredicate": {"PP": "3-ff00:0:3#0"}
    }
  },
  "go-through-frankfurt": {
    "ActionFilterPaths": {
      "CondPathPredicate": {"PP": "1-ff00:0:12#0"}
    }
  }
}
```

Now that the selectors are in place, define SIG sessions through Zürich and Frankfurt for the remote
London AS:

```json
"ASes": {
  "2-ff00:0:2": {
    "Nets": ["192.168.2.0/24", "192.168.3.0/24"],
    "Sigs": {
      "LondonA": {
        "Addr": "192.168.2.1"
      }
    },
    "Sessions": {
      "10": "go-through-zuerich",
      "20": "go-through-frankfurt"
    }
  }
}
```

With the sessions in place, we can define the policy:

```json
"PktPolicies": [
  {"Classname": "all", "SessIDs": [10, 20]}
]
```

The final configuration file for the Berlin SIG:

```json
{
  "ASes": {
    "2-ff00:0:2": {
      "Nets": ["192.168.2.0/24", "192.168.3.0/24"],
      "Sigs": {
        "LondonA": {
          "Addr": "192.168.2.1"
        }
      },
      "Sessions": {
        "10": "go-through-zuerich",
        "20": "go-through-frankfurt"
      },
      "PktPolicies": [
        {"Classname": "all", "SessIDs": [10, 20]}
      ]
    }
  },
  "Classes": {
    "all": {
      "CondBool": true
    }
  }
  "Actions": {
    "go-through-zuerich": {
      "ActionFilterPaths": {
        "CondPathPredicate": {"PP": "3-ff00:0:3#0"}
      }
    },
    "go-through-frankfurt": {
      "ActionFilterPaths": {
        "CondPathPredicate": {"PP": "1-ff00:0:12#0"}
      }
    }
  }
}
```

#### Choosing different paths for different traffic <a id="different-paths"> </a>

Suppose traffic going to `192.168.2.0/24` should prefer paths through Frankfurt, and only go through
Zürich if that path becomes unavailable. Also, traffic to `192.168.3.0/24` should prefer going
through Zürich, and only go through Frankfurt if that path becomes unavailable. In this case, the
path selectors and sessions from the previous section remain the same. However, we need to define
two new classes for the criteria we chose:

```json
"Classes": {
  "network-1-traffic": {
    "CondIPv4": { "MatchDestination": {"Net": "192.168.2.0/24"}}
  },
  "network-2-traffic": {
    "CondIPv4": { "MatchDestination": {"Net": "192.168.3.0/24"}}
  }
}
```

Now, define the policies that map the classes to sessions. Sessions 10 and 20 are the ones from the
previous section.

```json
"PktPolicies": [
  {"Classname": "network-1-traffic", "SessIDs": [20, 10]},
  {"Classname": "network-2-traffic", "SessIDs": [10, 20]}
]
```

For traffic going to `192.168.2.0/24`, we prefer session 20 (with paths through Frankfurt) and
choose 10 (with paths through Zürich) only if the former is unhealthy. For traffic going to
`192.168.3.0/24`, we prefer paths through Zürich, and then through Frankfurt.

#### Session of last resort <a id="session-of-last-resort"> </a>

One useful pattern is to always have a "session of last resort" that allows any path to be used. For
this, define a session without a selector (typically, session 0):

```json
"Sessions": {
  "10": "go-through-zuerich",
  "20": "go-through-frankfurt",
  "0": "",
}
```

This session can then be used at the end of policies to have traffic use any available path if the
preferred ones are no longer available:

```json
"PktPolicies": [
  {"Classname": "network-1-traffic", "SessIDs": [20, 10, 0]},
  {"Classname": "network-2-traffic", "SessIDs": [10, 20, 0]}
]
```

In this topology, session 0 will not provide any additional reliability because all the paths are
already used in sessions 10 and 20. However, in more complex deployments where the number of paths
is much greater a session of last resort can be useful to ensure that traffic continues to flow.
