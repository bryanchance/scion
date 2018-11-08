# Copyright 2018 Anapaya Systems

# SCION
from topology.ana.docker import DockerGenerator
from topology.ana.go import GoGenerator
from topology.ana.postgres import PostgresGenerator
from topology.ana.supervisor import SupervisorGenerator
from topology.config import ConfigGenerator as VanillaGenerator


class ConfigGenerator(VanillaGenerator):

    def _generate_go(self, topo_dicts):
        args = self._go_args(topo_dicts)
        go_gen = GoGenerator(args)
        if self.args.cert_server == "go":
            go_gen.generate_cs()
        if self.args.sciond == "go":
            go_gen.generate_sciond()
        if self.args.path_server == "go":
            go_gen.generate_ps()
        if self.args.discovery:
            go_gen.generate_ds()
        # XXX(lukedirtwalker): Optimally we would call this from the top level generator,
        # but to hide it from scionproto code we do it here.
        # TODO(lukedirtwalker): Add flag for backend: https://github.com/Anapaya/scion/issues/235
        self._generate_postgres()

    def _generate_docker(self, topo_dicts):
        args = self._docker_args(topo_dicts)
        docker_gen = DockerGenerator(args)
        docker_gen.generate()

    def _generate_supervisor(self, topo_dicts):
        args = self._supervisor_args(topo_dicts)
        super_gen = SupervisorGenerator(args)
        super_gen.generate()

    def _generate_postgres(self):
        pg_gen = PostgresGenerator(self.args.output_dir, self.args.in_docker)
        pg_gen.generate()
