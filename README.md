# Bond CoreDNS Plugin


## Overview
`bond` is **Bitcoin**, **Ordinals**, and **Nostr**, into **DNS**. It decentralizes domain name resolution using **Bitcoin** and **Ordinals** as the domain registration protocol, and [**Nostr**](https://github.com/nostr-protocol/nostr) for dynamic IP address lookup.

In practice, that means that `bond` resolves [`.btc` domain names](https://docs.btcname.id/docs) such as `godofthunder.btc` or `rowboto.btc` to IP addresses posted in **Nostr** notes.

> **Warning**: The current implementation is a **Proof of Concept** and is not production ready. Do not use it on mainnet **by any means**.


## CoreDNS Plugin
This plugin allows CoreDNS to resolve `.btc` domains to IPv4 addresses through a [`bond-index`](https://github.com/RowDaBoat/bond-index) server and Nostr.


## Installation
To install this plugin into CoreDNS you must compile them together.
1- clone CoreDNS
```bash
git clone https://github.com/coredns/coredns.git
cd coredns
```

2- Add this plugin at the end of the `plugins.cfg` file:
```
...
bond:github.com/RowDaBoat/bond-coredns
```

3- Get the plugin's source:
```bash
go get https://github.com/RowDaBoat/bond-coredns.git
```

4- Compile CoreDNS:
```bash
go generate
go build
```

5- Check bond is present:
```bash
./coredns -plugins
```

6- Start CoreDNS:
```bash
./coredns
```
`bond` should show up in the output.


## Configuration
Add the following to your `Corefile` configuration file:
```
btc.:53 {
    bond bond_server_ip:port
}
```
