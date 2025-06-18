# Router

- OS: OpenBSD

## Firewall
```pf
# Block by default
block drop log all keep state

# Allow any Outgoing traffic
pass out quick all keep state

# Incoming
pass in quick proto tcp from CIDR to IP port ssh keep state
```
