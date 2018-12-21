# Copyright 2018 Anapaya Systems

# Stdlib
import json
import os
# SCION
from lib.util import write_file
from topology.common import json_default
from topology.sig import SIGGenerator as VanillaGenerator


class SIGGenerator(VanillaGenerator):

    def _sig_json(self, topo_id):
        sig_cfg = {"ConfigVersion": 1, "ASes": {}, "Classes": {"default": {"CondBool": True}}}
        for t_id, topo in self.args.topo_dicts.items():
            if topo_id == t_id:
                continue
            sig_cfg['ASes'][str(t_id)] = {
                "Nets": [],
                "Sessions": {"0": ""},
                "PktPolicies": [{"ClassName": "default", "SessIds": [0]}]
            }
            net = self.args.networks['sig%s' % t_id.file_fmt()][0]
            sig_cfg['ASes'][str(t_id)]['Nets'].append(net['net'])

        cfg = os.path.join(topo_id.base_dir(self.args.output_dir), 'sig%s' % topo_id.file_fmt(),
                           "cfg.json")
        contents_json = json.dumps(sig_cfg, default=json_default, indent=2)
        write_file(cfg, contents_json + '\n')
