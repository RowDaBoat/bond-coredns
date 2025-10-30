package bond_coredns

import (
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("bond", setup) }

func setup(c *caddy.Controller) error {
	c.Next()

	// bond token
	if !c.NextArg() {
		return plugin.Error("bond", c.ArgErr())
	}

	var endpoint = c.Val()
	if c.NextArg() {
		return plugin.Error("bond", c.ArgErr())
	}

	if host, port, err := net.SplitHostPort(endpoint); err != nil || host == "" || port == "" {
		return plugin.Error("bond", c.Errf("invalid endpoint %q, expected host:port", endpoint))
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return BondCoreDns{Next: next, Endpoint: endpoint}
	})

	return nil
}
