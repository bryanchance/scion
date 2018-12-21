# Copyright 2018 Anapaya Systems

# Stdlib
import os
import toml

# SCION
from lib.util import write_file
from topology.common import docker_host, get_l4_port, get_pub, get_pub_ip
from topology.go import GoGenerator as VanillaGenerator
from topology.ana.postgres import CSDB_NAME, CSDB_PORT, PSDB_NAME, PSDB_PORT


DB_TYPE_TRUST = 'trust'
DB_TYPE_PATH = 'path'
SVC_PS = 'ps'
SVC_CS = 'cs'


class GoGenerator(VanillaGenerator):

    def generate_ds(self):
        for topo_id, topo in self.args.topo_dicts.items():
            acl = self._build_ds_acl(topo)
            for k, v in topo.get("DiscoveryService", {}).items():
                base = topo_id.base_dir(self.args.output_dir)
                ds_conf = self._build_ds_conf(topo_id, base, k, v)
                write_file(os.path.join(base, k, "acl"), acl)
                write_file(os.path.join(base, k, "dsconfig.toml"), toml.dumps(ds_conf))

    def _build_ds_acl(self, topo):
        acl = []
        for k, conf in topo["BorderRouters"].items():
            acl.append("%s/32" % get_pub(conf["InternalAddrs"])['PublicOverlay']['Addr'].ip)
        for svc in (
            "BeaconService",
            "DiscoveryService",
            "CertificateService",
            "PathService",
        ):
            for k, conf in topo[svc].items():
                acl.append("%s/32" % get_pub_ip(conf["Addrs"]))
        return "\n".join(acl)

    def _build_ds_conf(self, topo_id, base, name, conf):
        config_dir = '/share/conf' if self.args.docker else os.path.join(base, name)
        log_dir = '/share/logs' if self.args.docker else 'logs'
        raw_entry = {
            'general': {
                'ID': name,
                'ConfigDir': config_dir,
            },
            'logging': {
                'file': {
                    'Path': os.path.join(log_dir, "%s.log" % name),
                    'Level': 'debug',
                },
                'console': {
                    'Level': 'crit',
                },
            },
            'infra': {
                'Type': "DS"
            },
            'ds': {
                "IA": str(topo_id),
                "ACL": os.path.join(config_dir, "acl"),
                "UseFileModTime": True,
                "ListenAddr": self._get_laddr(conf["Addrs"]),
                "zoo": {
                    "Instances": self._get_zk_instances(topo_id)
                }
            },
        }
        return raw_entry

    def _get_zk_instances(self, topo_id):
        zk = []
        for info in self.args.topo_dicts[topo_id]["ZookeeperService"].values():
            zk.append("%s:%d" % (info["Addr"], info["L4Port"]))
        return zk

    def _get_laddr(self, addr):
        ip = get_pub_ip(addr)
        # Allow DS to be reachable through port forwarding.
        if self.args.docker:
            ip = "0.0.0.0"
        return "%s:%s" % (ip, get_l4_port(addr))

    def generate_ps(self):
        for topo_id, topo in self.args.topo_dicts.items():
            self._generate_path_postgres_init(topo_id, SVC_PS)
            self._generate_trust_postgres_init(topo_id, SVC_PS)
            base = topo_id.base_dir(self.args.output_dir)
            for k, v in topo.get("PathService", {}).items():
                # without consul only a single go PS is supported
                if k.endswith("-1") or self.args.consul:
                    ps_conf = self._build_ps_conf(topo_id, topo["ISD_AS"], base, k)
                    write_file(os.path.join(base, k, "psconfig.toml"), toml.dumps(ps_conf))
                    self._genereate_pg_user_init(topo_id, SVC_PS, DB_TYPE_PATH, k)
                    self._genereate_pg_user_init(topo_id, SVC_PS, DB_TYPE_TRUST, k)

    def _build_ps_conf(self, topo_id, ia, base, name):
        raw_entry = super()._build_ps_conf(topo_id, ia, base, name)
        if self.args.ps_db != 'postgres':
            return raw_entry
        db_host = docker_host(self.args.in_docker, self.args.docker, 'localhost')
        db_user = self._pg_db_user(SVC_PS, DB_TYPE_PATH, name)
        raw_entry['ps']['PathDB'] = {
            'Backend': 'postgres',
            # sslmode=disable is because dockerized postgres doesn't have SSL enabled.
            'Connection': 'host=%s user=%s password=password' % (db_host, db_user) +
                          ' sslmode=disable dbname=%s port=%s' % (PSDB_NAME, PSDB_PORT),
        }
        raw_entry['ps']['RevCache'] = {
            'Backend': 'postgres',
        }
        db_user = self._pg_db_user(SVC_PS, DB_TYPE_TRUST, name)
        raw_entry['TrustDB'] = {
            'Backend': 'postgres',
            'Connection': 'host=%s user=%s password=password' % (db_host, db_user) +
                          ' sslmode=disable dbname=%s port=%s' % (PSDB_NAME, PSDB_PORT),
        }
        return raw_entry

    def generate_cs(self):
        for topo_id, topo in self.args.topo_dicts.items():
            self._generate_trust_postgres_init(topo_id, SVC_CS)
            for k, v in topo.get("CertificateService", {}).items():
                # only a single Go-CS per AS is currently supported
                if k.endswith("-1"):
                    base = topo_id.base_dir(self.args.output_dir)
                    cs_conf = self._build_cs_conf(topo_id, topo["ISD_AS"], base, k)
                    write_file(os.path.join(base, k, "csconfig.toml"), toml.dumps(cs_conf))
                    self._genereate_pg_user_init(topo_id, SVC_CS, DB_TYPE_TRUST, k)

    def _build_cs_conf(self, topo_id, ia, base, name):
        raw_entry = super()._build_cs_conf(topo_id, ia, base, name)
        if self.args.cs_db != 'postgres':
            return raw_entry
        db_user = self._pg_db_user(SVC_CS, DB_TYPE_TRUST, name)
        db_host = docker_host(self.args.in_docker, self.args.docker, 'localhost')
        raw_entry['TrustDB'] = {
            'Backend': 'postgres',
            'Connection': 'host=%s user=%s password=password' % (db_host, db_user) +
                          ' sslmode=disable dbname=%s port=%s' % (CSDB_NAME, CSDB_PORT),
        }
        return raw_entry

    def _generate_path_postgres_init(self, topo_id, svc):
        if not self._pg_db_enabled(svc):
            return
        self._generate_postgres_init(topo_id, svc, DB_TYPE_PATH, 'go/path_srv/postgres/schema.sql')

    def _generate_trust_postgres_init(self, topo_id, svc):
        if not self._pg_db_enabled(svc):
            return
        self._generate_postgres_init(topo_id, svc, DB_TYPE_TRUST, 'go/cert_srv/postgres/schema.sql')

    def _generate_postgres_init(self, topo_id, svc, db_type, schema_file):
        schema_name = self._pg_db_schema(topo_id, svc, db_type)
        sql = 'CREATE SCHEMA "%s";\n' % schema_name
        sql += 'SET search_path TO "%s";\n' % schema_name
        with open(schema_file, 'r') as schema:
            sql += schema.read()
        write_file(self._pg_init_file_path(topo_id, svc, db_type), sql)

    def _genereate_pg_user_init(self, topo_id, svc, db_type, name):
        if not self._pg_db_enabled(svc):
            return
        schema_name = self._pg_db_schema(topo_id, svc, db_type)
        db_user = self._pg_db_user(svc, db_type, name)
        sql = 'CREATE USER "%s" WITH PASSWORD \'password\';\n' % db_user
        sql += 'GRANT ALL ON SCHEMA "%s" TO "%s";\n' % (schema_name, db_user)
        sql += 'GRANT ALL ON ALL TABLES IN SCHEMA "%s" TO "%s";\n' % (schema_name, db_user)
        sql += 'GRANT ALL ON ALL SEQUENCES IN SCHEMA "%s" TO "%s";\n' % (schema_name, db_user)
        sql += 'GRANT ALL ON ALL FUNCTIONS IN SCHEMA "%s" TO "%s";\n' % (schema_name, db_user)
        sql += 'ALTER ROLE "%s" SET search_path TO "%s";\n' % (db_user, schema_name)
        with open(self._pg_init_file_path(topo_id, svc, db_type), 'a') as init_file:
            init_file.write(sql)

    def _pg_init_file_path(self, topo_id, svc, db_type):
        return os.path.join(self.args.output_dir, 'postgres_%s' % svc,
                            'init', '%s.sql' % self._pg_db_schema(topo_id, svc, db_type))

    def _pg_db_user(self, svc, db_type, name):
        return '%s%s_%s' % (svc, db_type, name)

    def _pg_db_schema(self, topo_id, svc, db_type):
        return '%s%s_%s' % (svc, db_type, topo_id.file_fmt())

    def _pg_db_enabled(self, svc):
        return ((svc == SVC_PS and self.args.ps_db == 'postgres') or
                (svc == SVC_CS and self.args.cs_db == 'postgres'))
