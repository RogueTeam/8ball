# 8ball

Super Simple XMR payment gateway.

Create payment addresses that are easy to track and which forward funds to a secured wallet.

## Binaries

```shell
go build -o build/ ./cmd/...
```


- `http2socks`: just a small tooling for translating an HTTP proxy request to a SOCKS5.
- `gateway`: Payment gateway with commissions enabled. Only two endpoints, see: https://xmrgateway.com/
- `tunnel`: internal tool for connecting two machines securely without the pain of Let's encrypt automation. (Uses libp2p)
