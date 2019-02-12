# Common functionality for patroni tests.

. acceptance/common.sh

CONSUL1=consul_server1
CONSUL2=consul_server2
PATRONI1=patroni_server1
PATRONI2=patroni_server2

TEST_ARTIFACTS_DIR="${ACCEPTANCE_ARTIFACTS:?}/${TEST_NAME}/patroni_files"

cmd_dc() {
    BASE_DIR="${SCION_OUTPUT_BASE:+$SCION_OUTPUT_BASE/}$TEST_ARTIFACTS_DIR"\
    COMPOSE_FILE="acceptance/patroni/patroni-dc.yml" docker-compose -p acceptance_patroni --no-ansi "$@"
}

test_setup() {
    set -e
    # First build the patroni_dev image
    ./tools/quiet ./docker.sh patroni_dev
    # Copy files to output dir
    mkdir -p "$TEST_ARTIFACTS_DIR/initsql"
    cp acceptance/patroni/patroni*.yml "$TEST_ARTIFACTS_DIR"
    cp acceptance/patroni/setup_cluster "$TEST_ARTIFACTS_DIR"
    cp acceptance/patroni/initsql/*.sql "$TEST_ARTIFACTS_DIR/initsql/"
    # start containers
    cmd_dc up -d
    # make sure they are ready
    sleep 30
}

test_teardown() {
    collect_docker_logs "cmd_dc"
    cmd_dc down
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
