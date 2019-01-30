# Copyright 2019 Anapaya Systems

from topology.ana.postgres import (
    CSDB_NAME,
    CSDB_PORT,
    pg_db_user,
)
from topology.common import (
    docker_host,
    trust_db_conf_entry as trust_db_conf_entry_base,
)

DS_CONFIG_NAME = 'dsconfig.toml'

DB_TYPE_TRUST = 'trust'
DB_TYPE_PATH = 'path'
SVC_PS = 'ps'
SVC_CS = 'cs'


def trust_db_conf_entry(args, name):
    if args.cs_db != 'postgres':
        return trust_db_conf_entry_base(args, name)
    db_user = pg_db_user(SVC_CS, DB_TYPE_TRUST, name)
    db_host = docker_host(args.in_docker, args.docker, 'localhost')
    return {
        'Backend': 'postgres',
        'Connection': 'host=%s user=%s password=password' % (db_host, db_user) +
                      ' sslmode=disable dbname=%s port=%s' % (CSDB_NAME, CSDB_PORT),
    }
