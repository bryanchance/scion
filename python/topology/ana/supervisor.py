# Copyright 2018 Anapaya Systems

# Stdlib
import os

# SCION
from topology.supervisor import SupervisorGenerator as VanillaGenerator


class SupervisorGenerator(VanillaGenerator):

    def _as_conf(self, topo, base):
        entries = super()._as_conf(topo, base)
        entries.extend(self._ds_entries(topo, base))
        return entries

    def _ds_entries(self, topo, base):
        entries = []
        for k, v in topo.get("DiscoveryService", {}).items():
            conf = os.path.join(base, k, "dsconfig.toml")
            entries.append((k, ["bin/discovery", "-config", conf]))
        return entries
