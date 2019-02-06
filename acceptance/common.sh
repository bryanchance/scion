# Acceptance tests common functions

#######################################
# Converts ISD-AS representation to its file format
# Arguments:
#   ISD-AS
#######################################
ia_file() {
    echo ${1:?} | sed -e "s/:/_/g"
}

#######################################
# Converts ISD-AS representation to AS file format, i.e. removes the ISD
# Arguments:
#   ISD-AS
#######################################
as_file() {
    ia_file ${1:?} | cut -d '-' -f 2
}

#######################################
# Collects metrics from services that can be found in the topology file
# Metrics are saved in METRICS_DIR/
# Arguments:
#   Topology path
#   METRICS_DIR
#######################################
collect_metrics() {
    TOPOLOGY=${1:?}
    echo "Reading topology: $TOPOLOGY"
    METRICS_DIR=${2:?}
    echo "Saving metrics in $METRICS_DIR"
    mkdir -p "$METRICS_DIR"

    collect_elem_metrics 'BorderRouters' '30442' 'InternalAddrs.IPv4.PublicOverlay.Addr'
    collect_elem_metrics 'SIG' '30456'
    collect_elem_metrics 'BeaconService'
    collect_elem_metrics 'CertificateService'
    collect_elem_metrics 'PathService'
}

#######################################
# Collect metrics for one element type
# Arguments:
#   Service Type
#   Port
#   IP Addr Key (optional)
#######################################
collect_elem_metrics() {
    local elems=$(jq -r ".${1:?} | keys | .[]" $TOPOLOGY)
    local addr_key=${3:-Addrs.IPv4.Public.Addr}
    for elem in $elems; do
        local ip="$(jq -r .$1[\"$elem\"].$addr_key $TOPOLOGY)"
        local topo_port="$(jq -r .$1[\"$elem\"].Addrs.IPv4.Public.L4Port $TOPOLOGY)"
        echo "Collect $elem metrics from $ip:${2:-$topo_port}"
        curl "$ip:${2:-$topo_port}/metrics" -o "$METRICS_DIR/$elem" -s -S --connect-timeout 2 || true
    done
}

#######################################
# Print docker container status
#######################################
docker_status() {
    log "Docker containers:"
    docker ps -a -s
}

#######################################
# Generic test_teardown, prints docker status and stops all containers
#######################################
test_teardown() {
    docker_status
    ./tools/dc down
}

#######################################
# Log: Echo with a timestamp
#######################################
log() {
    echo "$(date -u +"%F %T.%6N%z") $@"
}

#######################################
# Return the ip of the container
# Arguments:
#   Name of the container
#######################################
container_ip() {
    docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$1"
}

#######################################
# Collect docker compose logs into logs/docker
# Arguments:
#   The docker compose bash method to call.
#######################################
collect_docker_logs() {
    local cmd="${1:?"Missing cmd argument"}"
    local out_dir=logs/docker
    mkdir -p "$out_dir"
    for svc in $("$cmd" config --services); do
        "$cmd" logs $svc &> $out_dir/$svc.log
    done
}
