# Copyright 2018 Anapaya Systems

# Stdlib
import copy
import os

# SCION
from topology.common import DOCKER_USR_VOL
from topology.docker import DockerGenerator as VanillaGenerator


class DockerGenerator(VanillaGenerator):

    def _gen_topo(self, topo_id, topo, base):
        super()._gen_topo(topo_id, topo, base)
        self._ds_conf(topo_id, topo, base)

    def _ds_conf(self, topo_id, topo, base):
        raw_entry = {
            'image': 'scion_discovery',
            'environment': {
                'SU_EXEC_USERSPEC': self.user_spec,
            },
            'networks': {},
            'volumes': [
                *DOCKER_USR_VOL,
                self._logs_vol(),
            ],
            'command': [],
        }
        for k, v in topo.get("DiscoveryService", {}).items():
            entry = copy.deepcopy(raw_entry)
            entry['container_name'] = self.prefix + k
            net = self.elem_networks[k][0]
            entry['networks'][self.bridges[net['net']]] = {'ipv4_address': str(net['ipv4'])}
            entry['volumes'].append('%s:/share/conf:ro' % os.path.join(base, k))
            self.dc_conf['services']['scion_%s' % k] = entry
