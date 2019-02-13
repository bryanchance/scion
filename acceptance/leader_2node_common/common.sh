# Common functionality for leader_2node tests. These are tests for the 2 node setup of consul.

# NOTE: currently this file assumes that ACCEPTANCE_ARTIFACTS is set.

. acceptance/common.sh

TEST_ARTIFACTS_DIR="${ACCEPTANCE_ARTIFACTS:?}/${TEST_NAME}"

current_leader_count() {
    id="${1:-*}"
    if [ ! -e "logs/${TEST_NAME}_1.log" ] && [ ! -e "logs/${TEST_NAME}_2.log" ]; then
        echo "0"
    else
        grep ISLEADER logs/"${TEST_NAME}"_${id}.log | grep -c ISLEADER
    fi
}

kill_consul_server() {
    log "Killing consul_server$1"
    cmd_dc kill "consul_server$1"
}

# Based on the given leader id, determine the follower id.
follower_id() {
    # make it easy to get the follower id, just use the leader id to index.
    local map=("X" "2" "1")
    echo "${map[$1]}"
}

leader_id() {
    grep -ohP "ISLEADER id=\d" logs/"$TEST_NAME"*.log | cut -d '=' -f 2
}

is_leader() {
    grep -q "ISLEADER id=$1" logs/"$TEST_NAME"*.log
}

check_no_overlapping_leader() {
    # Check history -> only one leader at a time.
    # For this we combine the logs and check that after each ISLEADER a LOSTLEADER is printed.
    # Optimally we would also check that the id matches the previous id.
    local cnt=0
    while read; do
        line="$REPLY"
        if [ "$((cnt % 2))" -eq 0 ]; then
            [[ "$line" == *ISLEADER* ]] || fail "FAIL: Expected ISLEADER on line $line"
        else
            [[ "$line" == *LOSTLEADER* ]] || fail "FAIL: Expected LOSTLEADER on line $line"
        fi
        : $((cnt+=1))
    done < <(grep logs/"$TEST_NAME"*.log LEADER)
    log "OK: No overlapping leader"
}

test_setup() {
    set -e
    log "Start consul servers"
    cmd_dc up -d
    wait_consul_cluster_init
    mkdir -p "$TEST_ARTIFACTS_DIR"
    ./bin/leader_acceptance -id 1 -agent "$(container_ip consul_server1):8500" -key foo/bar -noClusterLeader -log.console trace &> logs/"$TEST_NAME"_1.log &
    echo $! > "$TEST_ARTIFACTS_DIR/leader_acceptance_client1.pid"
    ./bin/leader_acceptance -id 2 -agent "$(container_ip consul_server2):8500" -key foo/bar -noClusterLeader -log.console trace &> logs/"$TEST_NAME"_2.log &
    echo $! > "$TEST_ARTIFACTS_DIR/leader_acceptance_client2.pid"
}

current_consul_leader_count() {
    docker logs consul_server1 2> /dev/null | grep -c "New leader elected"
}

wait_consul_cluster_init() {
    for i in $(seq 20); do
        [ "$(current_consul_leader_count)" -eq 1 ] && return
        sleep 1
    done
    fail "Error: Consul cluster not ready"
}

test_teardown() {
    kill "$(cat $TEST_ARTIFACTS_DIR/leader_acceptance_client1.pid)"
    kill "$(cat $TEST_ARTIFACTS_DIR/leader_acceptance_client2.pid)"
    collect_docker_logs "cmd_dc"
    cmd_dc down
}

cmd_dc() {
    COMPOSE_FILE="acceptance/leader_2node_common/consul-dc.yml" docker-compose -p acceptance_leader_2node --no-ansi "$@"
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
