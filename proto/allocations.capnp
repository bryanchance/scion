@0x9a1d79f5fc11a58f;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct Allocation {
    ia @0 :Text;
    networks @1 :List(Text);
}

struct AllocationsRequest {
    id @0 :Text;
}

struct AllocationsReply {
    id @0 :Text;
    allocations @1 :List(Allocation);
}
