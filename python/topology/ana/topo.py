# Copyright 2018 Anapaya Systems

from topology.topo import TopoGenerator as VanillaGenerator


class TopoGenerator(VanillaGenerator):

    def _srv_count(self, as_conf, conf_key, def_num):
        if conf_key == "path_servers" and self.args.path_server == "go" and self.args.consul:
            return as_conf.get(conf_key, def_num)
        return super()._srv_count(as_conf, conf_key, def_num)
