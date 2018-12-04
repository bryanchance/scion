# Copyright 2018 Anapaya Systems

# Stdlib
import os
import toml

# SCION
from lib.util import write_file
from topology.common import docker_host, get_l4_port, get_pub_ip
from topology.docker import DEFAULT_DOCKER_NETWORK
from topology.go import GoGenerator as VanillaGenerator


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
            acl.append("%s/32" % get_pub_ip(conf["CtrlAddr"]))
        # XXX(roosd): Allow border routers that run on the host.
        if self.args.docker:
            acl.append(DEFAULT_DOCKER_NETWORK)
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
            self._generate_ps_postgres_init(topo_id)
            for k, v in topo.get("PathService", {}).items():
                # only a single Go-PS per AS is currently supported
                if k.endswith("-1"):
                    base = topo_id.base_dir(self.args.output_dir)
                    ps_conf = self._build_ps_conf(topo_id, topo["ISD_AS"], base, k)
                    write_file(os.path.join(base, k, "psconfig.toml"), toml.dumps(ps_conf))

    def _build_ps_conf(self, topo_id, ia, base, name):
        raw_entry = super()._build_ps_conf(topo_id, ia, base, name)
        if self.args.path_db != "postgres":
            return raw_entry
        db_user = self._postgres_db_user(topo_id)
        db_host = docker_host(self.args.in_docker, self.args.docker, 'localhost')
        raw_entry['ps']['PathDB'] = {
            'Backend': 'postgres',
            # sslmode=disable is because dockerized postgres doesn't have SSL enabled.
            'Connection': 'host=%s user=%s password=password' % (db_host, db_user) +
                          ' sslmode=disable dbname=pathdb',
        }
        raw_entry['ps']['RevCache'] = {
            'Backend': 'postgres',
        }
        return raw_entry

    def _generate_ps_postgres_init(self, topo_id):
        if self.args.path_db != 'postgres':
            return
        db_user = self._postgres_db_user(topo_id)
        schema_name = db_user
        sql = 'CREATE USER "%s" WITH PASSWORD \'password\';\n' % db_user
        sql += 'ALTER ROLE "%s" SET search_path TO "%s";\n' % (db_user, schema_name)
        sql += 'CREATE SCHEMA "%s" AUTHORIZATION "%s";\n' % (schema_name, db_user)
        sql += 'SET search_path TO "%s";\n' % schema_name
        with open('go/path_srv/postgres/schema.sql', 'r') as schema:
            sql += schema.read()
        sql += '\nGRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA "%s" TO "%s";\n' % (schema_name,
                                                                                   db_user)
        write_file(os.path.join(self.args.output_dir, 'postgres', 'init', '%s.sql' % db_user), sql)

    def _postgres_db_user(self, topo_id):
        return topo_id.file_fmt()
