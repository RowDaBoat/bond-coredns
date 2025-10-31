package bond_coredns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("bond")

type BondCoreDns struct {
	Next     plugin.Handler
	Endpoint string
}

type NameResolution struct {
	NostrNpub   string   `json:"npub"`
	NostrRelays []string `json:"relays"`
}

type ResolvedIp struct {
	at  time.Time
	txt string
}

func (b BondCoreDns) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	qname, err := b.MsgToQname(r)
	if err != nil {
		return b.Finish(ctx, w, r)
	}

	nameResolution, err := b.NameToNostr(qname)
	if err != nil {
		return b.Finish(ctx, w, r)
	}

	ip, err := b.NostrToIP(nameResolution)
	if err != nil {
		return b.Finish(ctx, w, r)
	}

	return b.Reply(w, r, ip)
}

func (b BondCoreDns) Name() string { return "bond-coredns" }

type ResponsePrinter struct {
	dns.ResponseWriter
}

func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("Bond")
	return r.ResponseWriter.WriteMsg(res)
}

func (b BondCoreDns) MsgToQname(r *dns.Msg) (string, error) {
	if len(r.Question) == 0 {
		return "", fmt.Errorf("no question in message")
	}

	qname := r.Question[0].Name

	if strings.HasSuffix(qname, ".") {
		qname = qname[:len(qname)-1]
	}

	log.Info("Requested domain: " + qname)
	fmt.Println("Requested domain: ", qname)
	return qname, nil
}

func (b BondCoreDns) NameToNostr(qname string) (*NameResolution, error) {
	url := "http://" + b.Endpoint + "/name/" + url.PathEscape(qname)
	client := &http.Client{}
	resp, err := client.Get(url)

	if err != nil {
		log.Error("HTTP request failed: " + err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result NameResolution

	if err := json.Unmarshal(body, &result); err != nil {
		log.Error("Failed to parse name resolution JSON: " + err.Error())
		return nil, err
	}

	fmt.Println("npub:", result.NostrNpub)
	fmt.Println("relays:", strings.Join(result.NostrRelays, ","))
	return &result, nil
}

func (b BondCoreDns) NostrToIP(nameResolution *NameResolution) (net.IP, error) {
	t, data, err := nip19.Decode(nameResolution.NostrNpub)
	if err != nil || t != "npub" {
		log.Error("invalid npub: %v", err)
		return nil, fmt.Errorf("invalid npub: %v", err)
	}

	pubKey := data.(nostr.PubKey)
	limit := 10
	timeout := 5 * time.Second
	relays := nameResolution.NostrRelays

	filter := nostr.Filter{
		Authors: []nostr.PubKey{pubKey},
		Kinds:   []nostr.Kind{nostr.KindTextNote},
		Limit:   limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pool := nostr.NewPool(nostr.PoolOptions{})
	notes := pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})

	ip := ""
	var latest time.Time
	for note := range notes {
		at := time.Unix(int64(note.Event.CreatedAt), 0)
		if at.Before(latest) {
			continue
		}

		var parsed = net.ParseIP(note.Event.Content)
		if parsed == nil {
			continue
		}

		var ipv4 = parsed.To4()
		if ipv4 == nil {
			continue
		}

		ip = ipv4.String()
		latest = at
	}

	if ip == "" {
		var relays = strings.Join(relays, ",")
		var npub = nameResolution.NostrNpub
		log.Error("no valid IPv4 address found for npub: %s on relays %s", npub, relays)
		return nil, fmt.Errorf("no valid IPv4 address found for npub: %s on relays %s", npub, relays)
	}

	fmt.Println("resolved ip: " + ip)
	return net.ParseIP(ip), nil
}

func (b BondCoreDns) Reply(w dns.ResponseWriter, r *dns.Msg, ip net.IP) (int, error) {
	reply := new(dns.Msg)
	reply.SetReply(r)

	a := &dns.A{Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}, A: ip}
	reply.Answer = append(reply.Answer, a)

	if err := w.WriteMsg(reply); err != nil {
		log.Error("failed to write DNS response: " + err.Error())
		return dns.RcodeServerFailure, err
	}

	return dns.RcodeSuccess, nil
}

func (b BondCoreDns) Finish(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	pw := NewResponsePrinter(w)
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	return plugin.NextOrFailure(b.Name(), b.Next, ctx, pw, r)
}
