SCION HSR
=========
Implementation of the SCION High Speed Router (HSR) as a VPP plugin.

# FD.io deb repo
Instructions to add FD.io deb repository for 18.10 release:
<https://packagecloud.io/fdio/1810/install#manual-deb>

# Dependencies
You would need the following packages to build the SCION VPP plugin:
```
sudo apt install vpp-lib vpp-dev clang-6.0
```

Required packages for running VPP:
```
sudo apt install vpp
```

# Build steps
Just run make from the current c/hsr directory:
```
make
```

# Running steps
For the purpose of quickly testing the plugin, a modified startup.conf is required.
The easiest way is to stop the VPP service and run it manually with the provided startup.conf
in this repo that also provides an intereactive shell.

From this directory, run:
```
sudo vpp -c startup.conf
```

Once in the VPP shell:
```
exec testdata/udp
clear runtime
trace add pg-input 1
set interface ip scion-bypass loop0
packet-generator enable-stream
show trace
```

# Developer notes

## buffer size
VPP pre-allocates all the buffers used for packet processing, each buffer is 2048 Bytes.
There is always some free space at the start of the buffer, which by default is 128 Bytes.
That means that buffers are 128 + 2048 Bytes.

When jumbo frame support is enabled, multiple chained buffers are used per packet.
Current code does NOT support jumbo frames.

As we always expect ETH + IP + UDP + SCION, we are always going to be within the buffer.

## packet trace
There is no logging per packet in the traditional sense, as it is too expensive.
Errors are reported with counters and then there is packet trace on demand.

VPP process packets through a directed graph of nodes (for more details read you can read
about the internals in https://fdio-vpp.readthedocs.io/en/latest/index.html).
When tracing a packet, the packet processing is exactly the same except that each node creates
a trace for each packet which might contain packet data, usually the data relevant to the node,
but it can be anything that node considers necessary.

In the case of errors, the packet trace will likely finishing in the error-drop node, and its
trace will display the error. The trace of previous nodes might display bad data/values which
could be related to the error. Bottom line is, you will clearly see that something is off and
the error should help you to pinpoint the issue in the packet.
```
Packet 1

00:00:12:568984: pg-input
  stream udp, 100 bytes, 1 sw_if_index
  current data 0, length 100, free-list 0, clone-count 0, trace 0x0
  UDP: 192.168.1.2 -> 192.168.1.1
    tos 0x00, ttl 64, length 96, checksum 0xf739
    fragment id 0x0000
  UDP: 4321 -> 30046
    length 72, checksum 0x244d
00:00:12:569381: ip4-input
  UDP: 192.168.1.2 -> 192.168.1.1
    tos 0x00, ttl 64, length 96, checksum 0xf739
    fragment id 0x0000
  UDP: 4321 -> 30046
    length 72, checksum 0x244d
00:00:12:569389: ip4-scion-bypass
  intf-index 0, next 0, error 5
  UDP: 192.168.1.2 -> 192.168.1.1
    tos 0x00, ttl 64, length 96, checksum 0xf739
    fragment id 0x0000
  UDP: 4321 -> 30046
    length 72, checksum 0x244d
00:00:12:569397: error-drop
  ip4-scion-bypass: bad udp length
```
