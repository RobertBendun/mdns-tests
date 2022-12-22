# Testing service discovery

This repository is intended for testing mDNS service discovery.

1. Compile with `go build`
2. Launch server with `./services -listen -msg "hello sailor"`
3. Launch client to test if service has been discovered `./services`
