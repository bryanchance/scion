# Copyright 2018 Anapaya Systems

import os

import yaml

from lib.util import write_file
from topology.common import ArgsTopoDicts


PG_CONF = 'postgres-dc.yml'


class PostgresGenArgs(ArgsTopoDicts):
    pass


class PostgresGenerator(object):

    def __init__(self, args):
        """
        :param PostgresGenArgs args: Containes the passed command line arguments and topo dicts."
        """
        self.args = args
        self.pg_conf = {'version': '3', 'services': {}}
        self.output_base = os.environ.get('SCION_OUTPUT_BASE', os.getcwd())

    def generate(self):
        if self.args.path_server == 'py' or self.args.path_db != 'postgres':
            return
        name = 'postgres_docker' if self.args.in_docker else 'postgres'
        entry = {
            'image': 'postgres:latest',
            'container_name': name,
            'network_mode': 'bridge',
            'environment': {
                'POSTGRES_USER': 'pathdb',
                'POSTGRES_PASSWORD': 'password',
            },
            'volumes': [
                self.output_base + '/gen/postgres/init:/docker-entrypoint-initdb.d:ro'
            ],
            'ports': [
                '5432:5432'
            ]
        }
        self.pg_conf['services'][name] = entry
        write_file(os.path.join(self.args.output_dir, PG_CONF),
                   yaml.dump(self.pg_conf, default_flow_style=False))
