# Copyright 2018 Anapaya Systems

# Stdlib
import copy
import os

# SCION
from topology.common import get_l4_port, get_pub
from topology.docker import DockerGenerator as VanillaGenerator


class DockerGenerator(VanillaGenerator):

    def _gen_topo(self, topo_id, topo, base):
        super()._gen_topo(topo_id, topo, base)
        self._ds_conf(topo_id, topo, base)

    def _ds_conf(self, topo_id, topo, base):
        prefix = 'scion_docker_' if self.args.in_docker else 'scion_'
        raw_entry = {
            'image': 'scion_discovery',
            'depends_on': [
                'zookeeper'
            ],
            'environment': {
                'SU_EXEC_USERSPEC': self.user_spec,
            },
            'volumes': [
                '/etc/passwd:/etc/passwd:ro',
                '/etc/group:/etc/group:ro',
                self.output_base + '/logs:/share/logs:rw'
            ],
            'command': [],
        }
        for k, v in topo.get("DiscoveryService", {}).items():
            entry = copy.deepcopy(raw_entry)
            entry['container_name'] = prefix + k
            entry['volumes'].append('%s:/share/conf:ro' % os.path.join(base, k))
            ip = get_pub(v['Addrs'])["Public"]["Addr"].ip
            port = get_l4_port(v['Addrs'])
            entry['ports'] = ['%s:%d:%d' % (ip, port, port)]
            self.dc_conf['services']['scion_%s' % k] = entry
