# Common functionality for patroni tests.

. acceptance/common.sh

CONSUL1=consul_server1
CONSUL2=consul_server2
PATRONI1=patroni_server1
PATRONI2=patroni_server2

TEST_ARTIFACTS_DIR="${ACCEPTANCE_ARTIFACTS:?}/${TEST_NAME}"
PATRONI_FILES_DIR="$TEST_ARTIFACTS_DIR/patroni_files"

cmd_dc() {
    BASE_DIR="${SCION_OUTPUT_BASE:+$SCION_OUTPUT_BASE/}$PATRONI_FILES_DIR"\
    COMPOSE_FILE="acceptance/patroni/patroni-dc.yml" docker-compose -p acceptance_patroni --no-ansi "$@"
}

test_setup() {
    set -e
    # First build the patroni_dev image
    ./tools/quiet ./docker.sh patroni_dev
    # Copy files to output dir
    mkdir -p "$PATRONI_FILES_DIR/initsql"
    cp acceptance/patroni/patroni*.yml "$PATRONI_FILES_DIR"
    cp acceptance/patroni/setup_cluster "$PATRONI_FILES_DIR"
    cp acceptance/patroni/initsql/*.sql "$PATRONI_FILES_DIR/initsql/"
    # start containers
    cmd_dc up -d
    # make sure they are ready
    sleep 30
}

base_teardown() {
    collect_docker_logs "cmd_dc"
    cmd_dc down
}

# Returns information in the leader key in consul.
# Takes an optional modify index as argument to run this in blocking mode.
leader_info() {
    local idx=${1:-0}
    curl -sS "http://$consul1_ip:8500/v1/kv/service/ptest/leader?index=$idx"
}

# Extracts the name from the leader_info json block.
leader_name() {
    echo $1 | jq -r '.[0].Value' | base64 --decode
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
    shift 3
    case "$COMMAND" in
        name)
            echo $TEST_NAME ;;
        setup|run|teardown)
            "test_$COMMAND" "$@" ;;
        dc)
            cmd_dc "$@" ;;
        *) print_help $PROGRAM; exit 1 ;;
    esac
}
