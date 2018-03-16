@0xece076003323d351;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct PushACL {
    permit @0 :List(UInt64);
}
