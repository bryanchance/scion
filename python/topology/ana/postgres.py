# Copyright 2018 Anapaya Systems

import os

import yaml

from lib.util import write_file

PG_CONF = 'postgres-dc.yml'


class PostgresGenerator(object):

    def __init__(self, out_dir, in_docker):
        self.out_dir = out_dir
        self.in_docker = in_docker
        self.pg_conf = {'version': '3', 'services': {}}
        self.output_base = os.environ.get('SCION_OUTPUT_BASE', os.getcwd())

    def generate(self):
        name = 'postgres_docker' if self.in_docker else 'postgres'
        entry = {
            'image': 'postgres:latest',
            'container_name': name,
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
        write_file(os.path.join(self.out_dir, PG_CONF),
                   yaml.dump(self.pg_conf, default_flow_style=False))
