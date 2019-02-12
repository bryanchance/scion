#!/bin/bash

. acceptance/common.sh

BASE="discovery/v1"
STATIC="static"
DYNAMIC="dynamic"
REDUCED="reduced.json"
FULL="full.json"

PRIVILEGED=${PRIVILEGED:-"172.220.0.2"}
REGULAR=${REGULAR:-"172.220.0.1"}


base_add_addrs() {
    sudo -p "Setup virtual interfaces - [sudo] password for %p: " true
    sudo ip addr add "$PRIVILEGED/32" dev lo
    sudo ip addr add "$REGULAR/32" dev lo
}

base_scion_run() {
    ./scion.sh run nobuild
    # Allow slow services to register with consul.
    sleep 10
}

# Get status code when querying topology from discovery service.
#
# Arguments:
#   Local host info
#   Host info of discovery service
#   Static or dynamic
#   Full or reduced topology
query_status_code() {
    url="$2/$BASE/$3/$4"
    curl -sS --interface $1 $url -w "%{http_code}" -o /dev/null || fail "Error: Unable to fetch status code. addr=$url"
}

# Query topology from discovery service.
#
# Arguments:
#   Local host info
#   Host info of discovery service
#   Static or dynamic
#   Full or reduced topology
query_topo() {
    url="$2/$BASE/$3/$4"
    curl -sS --interface $1 $url || fail "Error: Unable to fetch topology. addr=$url"
}

test_teardown() {
    sudo -p "Teardown virtual interfaces - [sudo] password for %p: " true
    sudo ip addr del "$PRIVILEGED/32" dev lo
    sudo ip addr del "$REGULAR/32" dev lo

    ./tools/dc collect_logs consul logs/docker
    ./tools/dc down
}

print_help() {
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

PROGRAM=`basename "$0"`
COMMAND="$1"

do_command() {
    PROGRAM="$1"
    COMMAND="$2"
    TEST_NAME="$3"
    shift 3
    case "$COMMAND" in
        name)
            echo $TEST_NAME ;;
        setup|run|teardown)
            "test_$COMMAND" "$@" ;;
        *) print_help; exit 1 ;;
    esac
}
