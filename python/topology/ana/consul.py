# Copyright 2018 Anapaya Systems
import json
import os
from pathlib import Path

import toml
import yaml

from lib.util import write_file
from topology.ana.common import DS_CONFIG_NAME
from topology.common import (
    ArgsTopoDicts,
    CS_CONFIG_NAME,
    docker_host,
    get_pub_ip,
    PS_CONFIG_NAME,
    SIG_CONFIG_NAME,
)

CLIENT_DIR = 'consul'
SERVER_DIR = 'consul_server'
CONSUL_DC = 'consul-dc.yml'

CLIENT_BASE_PORT = 8505
CONSUL_VERSION = '1.3.1'
CFG_FILE = 'general.json'


class ConsulGenArgs(ArgsTopoDicts):
    pass


class ConsulGenerator(object):

    def __init__(self, args):
        self.args = args
        self.dc_conf = {'version': '3', 'services': {}}
        self.output_base = os.environ.get('SCION_OUTPUT_BASE', os.getcwd())
        self.docker_ip = docker_host(self.args.in_docker, self.args.docker)

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
            'image': 'consul:%s' % CONSUL_VERSION,
            'container_name': 'consul_server_docker' if self.args.in_docker else 'consul_server',
            'network_mode': 'bridge',
            'ports': [
                '8301:8301',  # serf_lan
                '8302:8302',  # serf_wan
                '8500:8500',  # http
            ],
            'volumes': [
                '%s:/consul/cfg:ro' % os.path.join(self.output_base, self.args.output_dir,
                                                   SERVER_DIR)
            ],
            # First args are from dockerfile. We need to overwrite because of custom cfg dir.
            'command': ['agent', '-dev', '-client', '0.0.0.0', '-config-dir=/consul/cfg']
        }
        self.dc_conf['services']['consul_server'] = entry

    def _generate_server_general(self):
        conf = self._generate_consul_config('server', self.docker_ip)
        conf.update({
            'server': True,
            'ui': True,
        })
        write_file(os.path.join(self.args.output_dir, SERVER_DIR, CFG_FILE),
                   json.dumps(conf, indent=4))

    def _generate_client_agents(self):
        port = CLIENT_BASE_PORT
        for topo_id, topo in self.args.topo_dicts.items():
            base = topo_id.base_dir(self.args.output_dir)
            self._generate_client_dc(topo_id, topo, base, port)
            self._generate_client_config(topo_id, topo, base, port)
            port = port + 1

    def _generate_client_dc(self, topo_id, topo, base, port):
        name_pref = 'consul_agent_docker' if self.args.in_docker else 'consul_agent'
        entry = {
            'image': 'consul:%s' % CONSUL_VERSION,
            'container_name': '%s-%s' % (name_pref, topo_id.file_fmt()),
            'network_mode': 'bridge',
            'ports': [
                '%s:%s' % (port, port)
            ],
            'volumes': [
                '%s:/consul/cfg:ro' % os.path.join(self.output_base, base, CLIENT_DIR)
            ],
            'command': ['agent', '-dev', '-client', '0.0.0.0', '-config-dir=/consul/cfg']
        }
        self.dc_conf['services']['consul_agent-%s' % topo_id.file_fmt()] = entry

    def _generate_client_config(self, topo_id, topo, base, port):
        # Only one agent is created per AS in the test topology.
        for k, v in topo.get("CertificateService", {}).items():
            ip = str(get_pub_ip(v['Addrs']))
            break
        self._generate_client_general(topo_id, base, ip, port)
        self._generate_client_services(topo_id, topo, base, port)

    def _generate_client_general(self, topo_id, base, ip, port):
        conf = self._generate_consul_config('consul_agent-%s' % topo_id, ip)
        conf.update({
            'leave_on_terminate': True,
            'retry_join': [self.docker_ip],
        })
        conf['ports']['http'] = port
        write_file(os.path.join(base, CLIENT_DIR, CFG_FILE), json.dumps(conf, indent=4))

    def _generate_consul_config(self, name, ip):
        return {
            'datacenter': 'scion',
            'node_name': name,
            'server': False,
            "client_addr": ip,
            "bind_addr": '0.0.0.0',
            # Contrary to what the doc suggests, the addresses
            # do not default to client_addr.
            "addresses": {
                "http": '0.0.0.0',
            },
            # Default ports are also a lie.
            "ports": {
                "dns": -1,
                "grpc": -1,
            }
        }

    def _generate_client_services(self, topo_id, topo, base, port):
        services = {'BeaconService': 'bsconfig.toml',
                    'CertificateService': CS_CONFIG_NAME,
                    'DiscoveryService': DS_CONFIG_NAME,
                    'PathService': PS_CONFIG_NAME}
        if self.args.sig:
            services['SIG'] = SIG_CONFIG_NAME
        for svc, conf in services.items():
            for elem_id, v in topo.get(svc, {}).items():
                self._add_consul_to_conf(topo_id, base, elem_id, conf, port)

    def _add_consul_to_conf(self, topo_id, base, elem_id, conf, port):
        file = os.path.join(base, elem_id, conf)
        if not Path(file).is_file():
            return
        with open(file, 'a') as f:
            f.write('\n')
            consul_entry = {
                'consul': {
                    'Enabled': True,
                    'Prefix': str(topo_id),
                    'Agent': '%s:%s' % (self.docker_ip, port),
                },
            }
            f.write(toml.dumps(consul_entry))
