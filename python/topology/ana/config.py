# Copyright 2018 Anapaya Systems

# SCION
from topology.ana.docker import DockerGenerator
from topology.ana.go import GoGenerator
from topology.ana.postgres import PostgresGenerator
from topology.ana.supervisor import SupervisorGenerator
from topology.config import ConfigGenerator as VanillaGenerator


class ConfigGenerator(VanillaGenerator):

    def _generate_go(self, topo_dicts):
        go_gen = GoGenerator(self.out_dir, topo_dicts, self.docker, self.in_docker)
        if self.sd == "go":
            go_gen.generate_sciond()
        if self.ps == "go":
            go_gen.generate_ps()
        if self.ds:
            go_gen.generate_ds()
        # XXX(lukedirtwalker): Optimally we would call this from the top level generator,
        # but to hide it from scionproto code we do it here.
        # TODO(lukedirtwalker): Add flag for backend: https://github.com/Anapaya/scion/issues/235
        self._generate_postgres()

    def _generate_docker(self, topo_dicts):
        docker_gen = DockerGenerator(
            self.out_dir, topo_dicts, self.sd, self.ps)
        docker_gen.generate()

    def _generate_supervisor(self, topo_dicts):
        super_gen = SupervisorGenerator(
            self.out_dir, topo_dicts, self.mininet, self.cs, self.sd, self.ps)
        super_gen.generate()

    def _generate_postgres(self):
        pg_gen = PostgresGenerator(self.out_dir, self.in_docker)
        pg_gen.generate()
