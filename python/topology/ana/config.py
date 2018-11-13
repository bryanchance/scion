# Copyright 2018 Anapaya Systems

# StdLib
import logging
import sys

# SCION
from topology.ana.consul import ConsulGenArgs, ConsulGenerator
from topology.ana.docker import DockerGenerator
from topology.ana.go import GoGenerator
from topology.ana.postgres import PostgresGenArgs, PostgresGenerator
from topology.ana.supervisor import SupervisorGenerator
from topology.config import ConfigGenerator as VanillaGenerator


class ConfigGenerator(VanillaGenerator):

    def __init__(self, args):
        super().__init__(args)
        if self.args.consul and self.args.docker:
            logging.critical("Currently cannot use consul with docker!")
            sys.exit(1)

    def _generate_with_topo(self, topo_dicts):
        super()._generate_with_topo(topo_dicts)
        self._generate_postgres(topo_dicts)
        self._generate_consul(topo_dicts)

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

    def _generate_docker(self, topo_dicts):
        args = self._docker_args(topo_dicts)
        docker_gen = DockerGenerator(args)
        docker_gen.generate()

    def _generate_supervisor(self, topo_dicts):
        args = self._supervisor_args(topo_dicts)
        super_gen = SupervisorGenerator(args)
        super_gen.generate()

    def _postgres_args(self, topo_dicts):
        return PostgresGenArgs(self.args, topo_dicts, self.port_gen)

    def _generate_postgres(self, topo_dicts):
        args = self._postgres_args(topo_dicts)
        pg_gen = PostgresGenerator(args)
        pg_gen.generate()

    def _generate_consul(self, topo_dicts):
        consul_gen = ConsulGenerator(ConsulGenArgs(self.args, topo_dicts))
        consul_gen.generate()
