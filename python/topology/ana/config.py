# Copyright 2018 Anapaya Systems

# SCION
from topology.ana.docker import DockerGenerator
from topology.ana.go import GoGenerator
from topology.ana.supervisor import SupervisorGenerator
from topology.config import ConfigGenerator as VanillaGenerator


class ConfigGenerator(VanillaGenerator):

    def _generate_go(self, topo_dicts):
        go_gen = GoGenerator(self.out_dir, topo_dicts, self.docker)
        if self.sd == "go":
            go_gen.generate_sciond()
        if self.ps == "go":
            go_gen.generate_ps()
        if self.ds:
            go_gen.generate_ds()

    def _generate_docker(self, topo_dicts):
        docker_gen = DockerGenerator(
            self.out_dir, topo_dicts, self.sd, self.ps)
        docker_gen.generate()

    def _generate_supervisor(self, topo_dicts):
        super_gen = SupervisorGenerator(
            self.out_dir, topo_dicts, self.mininet, self.cs, self.sd, self.ps)
        super_gen.generate()
