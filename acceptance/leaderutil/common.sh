# Copyright 2018 Anapaya Systems

# Common functionality for leader elcection tests.

. acceptance/common.sh

SERVER_NAME1=consul_server1
SERVER_NAME2=consul_server2
AGENT_NAME1=consul_agent1
AGENT_NAME2=consul_agent2

is_leader() {
    grep -q "ISLEADER id=$1" logs/*.log
}

lost_leader() {
    grep -q "LOSTLEADER id=$1" logs/*.log
}

cmd_dc() {
    COMPOSE_FILE="acceptance/leaderutil/consul-dc.yml" docker-compose -p acceptance_consul --no-ansi "$@"
}

test_setup() {
    set -e
    cmd_dc up -d
    sleep 2 # To make sure the system is ready to accept clients.
}

kill_consul() {
    cmd_dc down || true
}

print_help() {
    PROGRAM="$1"
    echo
	cat <<-_EOF
	    $PROGRAM name
	        return the name of this test
	    $PROGRAM setup
	        execute only the setup phase.
	    $PROGRAM run
	        execute only the run phase.
	    $PROGRAM teardown
	        execute only the teardown phase.
	_EOF
}

do_command() {
    PROGRAM="$1"
    COMMAND="$2"
    TEST_NAME="$3"
    case "$COMMAND" in
        name)
            echo $TEST_NAME ;;
        setup|run|teardown)
            "test_$COMMAND" ${@:4} ;;
        *) print_help $PROGRAM; exit 1 ;;
    esac
}
