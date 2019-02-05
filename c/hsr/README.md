SCION HSR
=========
Implementation of the SCION High Speed Router (HSR) as a VPP plugin.

# FD.io deb repo
Instructions to add FD.io deb repository for 18.10 release:
<https://packagecloud.io/fdio/1810/install#manual-deb>

# Dependencies
You would need the following packages to build the SCION VPP plugin:
```
sudo apt install vpp-lib vpp-dev
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
clear runtime
trace add pg-input 1
set interface ip scion-bypass loop0
packet-generator enable-stream
show trace
```

