# Copyright 2018 Anapaya Systems

# Stdlib
import os
import toml

# SCION
from lib.util import write_file
from topology.common import _get_l4_port, _get_pub_ip
from topology.docker import DEFAULT_DOCKER_NETWORK
from topology.go import GoGenerator as VanillaGenerator


class GoGenerator(VanillaGenerator):

    def generate_ds(self):
        for topo_id, topo in self.topo_dicts.items():
            acl = self._build_ds_acl(topo)
            for k, v in topo.get("DiscoveryService", {}).items():
                base = topo_id.base_dir(self.out_dir)
                ds_conf = self._build_ds_conf(topo_id, base, k, v)
                write_file(os.path.join(base, k, "acl"), acl)
                write_file(os.path.join(base, k, "dsconfig.toml"), toml.dumps(ds_conf))

    def _build_ds_acl(self, topo):
        acl = []
        for k, conf in topo["BorderRouters"].items():
            acl.append("%s/32" % _get_pub_ip(conf["CtrlAddr"]))
        # XXX(roosd): Allow border routers that run on the host.
        if self.docker:
            acl.append(DEFAULT_DOCKER_NETWORK)
        for svc in (
            "BeaconService",
            "DiscoveryService",
            "CertificateService",
            "PathService",
        ):
            for k, conf in topo[svc].items():
                acl.append("%s/32" % _get_pub_ip(conf["Addrs"]))
        return "\n".join(acl)

    def _build_ds_conf(self, topo_id, base, name, conf):
        config_dir = '/share/conf' if self.docker else os.path.join(base, name)
        log_dir = '/share/logs' if self.docker else 'logs'
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
        for info in self.topo_dicts[topo_id]["ZookeeperService"].values():
            zk.append("%s:%d" % (info["Addr"], info["L4Port"]))
        return zk

    def _get_laddr(self, addr):
        ip = _get_pub_ip(addr)
        # Allow DS to be reachable through port forwarding.
        if self.docker:
            ip = "0.0.0.0"
        return "%s:%s" % (ip, _get_l4_port(addr))
