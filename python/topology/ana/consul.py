# Copyright 2018 Anapaya Systems
import json
import os
from pathlib import Path

import toml
import yaml

from lib.util import write_file
from topology.common import get_l4_port, get_pub_ip, ArgsTopoDicts

CLIENT_DIR = 'consul'
SERVER_DIR = 'consul_server'
CONSUL_DC = 'consul-dc.yml'

SERVER_IP = '127.0.0.1'


class ConsulGenArgs(ArgsTopoDicts):
    pass


class ConsulGenerator(object):

    def __init__(self, args):
        self.args = args
        self.dc_conf = {'version': '3', 'services': {}}
        self.output_base = os.environ.get('SCION_OUTPUT_BASE', os.getcwd())

    def generate(self):
        if not self.args.consul:
            return
        self._generate_server_agents()
        self._generate_client_agents()
        write_file(os.path.join(self.args.output_dir, CONSUL_DC),
                   yaml.dump(self.dc_conf, default_flow_style=False))

    def _generate_server_agents(self):
        self._generate_server_dc()
        self._generate_server_general()

    def _generate_server_dc(self):
        entry = {
            'image': 'consul:latest',
            'container_name': self._server_name(),
            'network_mode': 'host',
            'volumes': [
                '%s:/consul/config' % os.path.join(self.output_base,
                                                   self.args.output_dir, SERVER_DIR)
            ]
        }
        self.dc_conf['services'][self._server_name()] = entry

    def _generate_server_general(self):
        conf = self._generate_consul_config('server', SERVER_IP)
        conf.update({'server': True})
        write_file(os.path.join(self.args.output_dir, SERVER_DIR, 'general.json'),
                   json.dumps(conf, indent=4))

    def _generate_client_agents(self):
        for topo_id, topo in self.args.topo_dicts.items():
            base = os.path.join(self.output_base, topo_id.base_dir(self.args.output_dir))
            self._generate_client_dc(topo_id, topo, base)
            self._generate_client_config(topo_id, topo, base)

    def _generate_client_dc(self, topo_id, topo, base):
        entry = {
            'image': 'consul:latest',
            'container_name': self._agent_name(topo_id.file_fmt()),
            'depends_on': [
                self._server_name(),
            ],
            'network_mode': 'host',
            'volumes': [
                '%s:/consul/config' % os.path.join(base, CLIENT_DIR)
            ],
        }
        self.dc_conf['services'][self._agent_name(topo_id.file_fmt())] = entry

    def _generate_client_config(self, topo_id, topo, base):
        # Only one agent is created per AS in the test topology.
        for k, v in topo.get("CertificateService", {}).items():
            ip = str(get_pub_ip(v['Addrs']))
            break
        self._generate_client_general(topo_id, base, ip)
        self._generate_client_services(topo_id, topo, base, ip)

    def _generate_client_general(self, topo_id, base, ip):
        conf = self._generate_consul_config(self._agent_name(topo_id), ip)
        conf.update({
            'leave_on_terminate': True,
            'retry_join': [SERVER_IP],
        })
        write_file(os.path.join(base, CLIENT_DIR, 'general.json'),
                   json.dumps(conf, indent=4))

    def _generate_consul_config(self, name, ip):
        return {
            'datacenter': 'scion',
            'ui': True,
            'node_name': name,
            "client_addr": ip,
            "bind_addr": ip,
            # Contrary to what the doc suggests, the addresses
            # do not default to client_addr.
            "addresses": {
                "http": ip,
            },
            # Default ports are also a lie.
            "ports": {
                "dns": -1,
                "grpc": -1,
            }
        }

    def _generate_client_services(self, topo_id, topo, base, agent_ip):
        for svc, conf in {'BeaconService': 'bsconfig.toml',
                          'CertificateService': 'csconfig.toml',
                          'PathService': 'psconfig.toml'}.items():
            for elem_id, v in topo.get(svc, {}).items():
                self._generate_client_svc(topo_id, topo, base, svc, elem_id, v)
                self._add_consul_to_conf(base, elem_id, conf, agent_ip)

    def _generate_client_svc(self, topo_id, topo, base, svc, elem_id, v):
        checks = [{
            'id': elem_id,
            'name': 'Health Check: %s' % elem_id,
            'ttl': '10s',
        }]
        # FIXME(roosd): BeaconService does not support consul health check yet.
        if svc == 'BeaconService':
            checks = []
        svc = {
            'services': [{
                'name': '%s/%s' % (topo_id, svc),
                'id': elem_id,
                'address': str(get_pub_ip(v['Addrs'])),
                'port': get_l4_port(v['Addrs']),
                'checks': checks,
            }]
        }
        write_file(os.path.join(base, CLIENT_DIR, '%s.json' % elem_id),
                   json.dumps(svc, indent=4))

    def _add_consul_to_conf(self, base, elem_id, conf, agent_ip):
        file = os.path.join(base, elem_id, conf)
        if not Path(file).is_file():
            return
        with open(file, 'a') as f:
            consul_entry = {
                'consul': {
                    'UpdateTTL': True,
                    'Agent': '%s:8500' % agent_ip,
                },
            }
            f.write(toml.dumps(consul_entry))

    def _server_name(self):
        return 'consul_server_docker' if self.args.in_docker else 'consul_server'

    def _agent_name(self, name):
        return 'agent-%s' % name
