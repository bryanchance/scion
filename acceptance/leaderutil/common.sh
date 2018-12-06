# Copyright 2018 Anapaya Systems

# Common functionality for leader elcection tests.

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

ip_of() {
    echo `docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$1"`
}

log() {
    echo "$(date -u --rfc-3339=seconds) $@"
}

test_setup() {
    set -e
    docker run -d --name="$SERVER_NAME1" -e 'CONSUL_LOCAL_CONFIG={"session_ttl_min": "1s"}' \
                                         -e CONSUL_BIND_INTERFACE=eth0 consul
    docker run -d --name="$SERVER_NAME2" -e 'CONSUL_LOCAL_CONFIG={"session_ttl_min": "1s"}' \
                                         -e CONSUL_BIND_INTERFACE=eth0 consul agent -server \
                                         -join="$(ip_of $SERVER_NAME1)"
    docker run -d --name="$AGENT_NAME1" -e 'CONSUL_LOCAL_CONFIG={"session_ttl_min": "1s"}' \
                                        -e CONSUL_CLIENT_INTERFACE=eth0 \
                                        -e CONSUL_BIND_INTERFACE=eth0 consul agent \
                                        -join="$(ip_of $SERVER_NAME1)"
    docker run -d --name="$AGENT_NAME2" -e 'CONSUL_LOCAL_CONFIG={"session_ttl_min": "1s"}' \
                                        -e CONSUL_CLIENT_INTERFACE=eth0 \
                                        -e CONSUL_BIND_INTERFACE=eth0 consul agent \
                                        -join="$(ip_of $SERVER_NAME2)"
    sleep 5 # To make sure the system is ready to accept clients.
}

kill_consul() {
    docker rm -f "$SERVER_NAME1" "$SERVER_NAME2" "$AGENT_NAME1" "$AGENT_NAME2" || true
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
