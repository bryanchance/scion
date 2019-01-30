# Copyright 2018 Anapaya Systems

import os

import yaml

from lib.util import write_file
from topology.common import ArgsTopoDicts


PG_CONF = 'postgres-dc.yml'
PSDB_NAME = 'psdb'
CSDB_NAME = 'csdb'
PSDB_PORT = 5432
CSDB_PORT = 5433


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
        if self.args.cs_db != 'postgres' and self.args.ps_db != 'postgres':
            return
        if self.args.ps_db == 'postgres' and self.args.path_server == 'go':
            self._gen_dc('postgres_ps', PSDB_NAME, PSDB_PORT)
        if self.args.cs_db == 'postgres' and self.args.cert_server == 'go':
            self._gen_dc('postgres_cs', CSDB_NAME, CSDB_PORT)
        write_file(os.path.join(self.args.output_dir, PG_CONF),
                   yaml.dump(self.pg_conf, default_flow_style=False))

    def _gen_dc(self, name_prefix, user, exp_port):
        name = '%s_docker' % name_prefix if self.args.in_docker else name_prefix
        entry = {
            'image': 'postgres:10',
            'container_name': name,
            'network_mode': 'bridge',
            'environment': {
                'POSTGRES_USER': user,
                'POSTGRES_PASSWORD': 'password',
            },
            'volumes': [
                self.output_base + '/gen/%s/init:/docker-entrypoint-initdb.d:ro' % name_prefix
            ],
            'ports': [
                '%s:5432' % exp_port
            ],
            'command': ['postgres', '-c', 'max_connections=200'],
        }
        self.pg_conf['services'][name_prefix] = entry


def pg_db_user(svc, db_type, name):
    return '%s%s_%s' % (svc, db_type, name)
