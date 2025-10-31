# Bond CoreDNS Plugin

This CoreDNS plugin allows resolving `.btc`` domains to IPv4 addresses through Nostr through a Bond server.

## Configuration
```
btc.:53 {
    bond bond_server_ip:port
}
```
