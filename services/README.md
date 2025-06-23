# Router

- OS: OpenBSD

## Firewall

```pf
set block-policy drop
set skip on lo

# Block by default
block drop log all

# Allowed services
pass in quick proto tcp from 10.20.30.0/24 to 10.20.30.200 port ssh  keep state
pass in quick proto tcp from 20.30.40.0/24 to 20.30.40.1   port 9050 keep state
pass in quick proto tcp from 20.30.40.0/24 to 20.30.40.1   port 8080 keep state

# Outgoing
pass out quick log all keep state
```
