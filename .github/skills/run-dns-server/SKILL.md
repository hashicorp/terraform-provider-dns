---
name: run-dns-server
description: Guide for running a local containerized DNS server. Use this when asked to run an acceptance test or to run a test with the prefix `TestAcc`.
---

Before running any acceptance tests for the DNS provider, start a DNS server:

1. Check whether the DNS server is already alive by querying its locally-mapped port:
   ```
   dig @127.0.0.1 -p 15353 +short ns.example.com
   ```
   Expected command output is "127.0.0.1."

1. If the DNS server is not already alive, then start the DNS server by running this command in the background:
   ```
   docker run --privileged --cgroupns=host -d --tmpfs /tmp --tmpfs /run \
    -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
    -v /etc/localtime:/etc/localtime:ro \
    -v $PWD/internal/provider/testdata/named.conf.kerberos:/etc/named.conf:ro \
    -p 127.0.0.1:15353:53 \
    -p 127.0.0.1:15353:53/udp \
    --rm --name ns --hostname ns.example.com ns
    ```
1. After starting the DNS server, verify that the DNS serer is alive by querying its locally-mapped port:
   ```
   dig @127.0.0.1 -p 15353 +short ns.example.com
   ```
   Expected command output is "127.0.0.1." Display no output unless there is a problem.

1. If there is a problem, do not run the acceptance tests.

Add these environment variables to all `go test` commands for acceptance tests in the DNS provider:
- `DNS_UPDATE_SERVER=127.0.0.1`
- `DNS_UPDATE_PORT=15353`
