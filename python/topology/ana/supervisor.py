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

    def _ps_entries(self, topo, base):
        if self.args.path_server == "py" or not self.args.consul:
            return super()._ps_entries(topo, base)
        entries = []
        for k, v in topo.get("PathService", {}).items():
            conf = os.path.join(base, k, "psconfig.toml")
            entries.append((k, ["bin/path_srv", "-config", conf]))
        return entries
