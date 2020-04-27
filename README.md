# tracert-as
tracert-as is a Go command line utility that combines traceroute and whois,
displaying AS-related whois information for every host on the route.

## Overview
For each host, the program attempts to query whois servers for the network name,
Autonomous System number, and country. For local addresses "local" is displayed instead.

## Privileges
tracert-as uses raw sockets since unprivileged sockets didn't return TTL exceeded packets
on my machine for some reason.  
This means you have to either run the program as root or give it raw network access privileges:
```sh
sudo setcap cap_net_raw+ep tracert-as
```

## Example usage
![Screenshot of the program's output](https://raw.githubusercontent.com/finkrer/tracert-as/master/examples/output.png)