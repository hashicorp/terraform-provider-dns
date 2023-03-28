package provider

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/miekg/dns"
)

const (
	defaultPort      = 53
	defaultRetries   = 3
	defaultTimeout   = "0"
	defaultTransport = "udp"
)

// New returns a *schema.Provider for DNS dynamic updates.
func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"update": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "When the provider is used for DNS updates, this block is required.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_SERVER", nil),
							Description: "The hostname or IP address of the DNS server to send updates to.",
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								if envPortStr := os.Getenv("DNS_UPDATE_PORT"); envPortStr != "" {
									port, err := strconv.Atoi(envPortStr)
									if err != nil {
										err = fmt.Errorf("invalid DNS_UPDATE_PORT environment variable: %s", err)
									}
									return port, err
								}

								return defaultPort, nil
							},
							Description: "The target UDP port on the server where updates are sent to. Defaults to `53`.",
						},
						"transport": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_TRANSPORT", defaultTransport),
							Description: "Transport to use for DNS queries. Valid values are `udp`, `udp4`, `udp6`, " +
								"`tcp`, `tcp4`, or `tcp6`. Any UDP transport will retry automatically with the " +
								"equivalent TCP transport in the event of a truncated response. Defaults to `udp`.",
						},
						"timeout": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_TIMEOUT", defaultTimeout),
							Description: "Timeout for DNS queries. Valid values are durations expressed as `500ms`, " +
								"etc. or a plain number which is treated as whole seconds.",
						},
						"retries": {
							Type:     schema.TypeInt,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								if env := os.Getenv("DNS_UPDATE_RETRIES"); env != "" {
									retries, err := strconv.Atoi(env)
									if err != nil {
										err = fmt.Errorf("invalid DNS_UPDATE_RETRIES environment variable: %s", err)
									}
									return retries, err
								}

								return defaultRetries, nil
							},
							Description: "How many times to retry on connection timeout. Defaults to `3`.",
						},
						"key_name": {
							Type:          schema.TypeString,
							Optional:      true,
							DefaultFunc:   schema.EnvDefaultFunc("DNS_UPDATE_KEYNAME", nil),
							ConflictsWith: []string{"update.0.gssapi.0"},
							RequiredWith:  []string{"update.0.key_algorithm", "update.0.key_secret"},
							Description:   "The name of the TSIG key used to sign the DNS update messages.",
						},
						"key_algorithm": {
							Type:          schema.TypeString,
							Optional:      true,
							DefaultFunc:   schema.EnvDefaultFunc("DNS_UPDATE_KEYALGORITHM", nil),
							ConflictsWith: []string{"update.0.gssapi.0"},
							RequiredWith:  []string{"update.0.key_name", "update.0.key_secret"},
							Description: "Required if `key_name` is set. When using TSIG authentication, the " +
								"algorithm to use for HMAC. Valid values are `hmac-md5`, `hmac-sha1`, `hmac-sha256` " +
								"or `hmac-sha512`.",
						},
						"key_secret": {
							Type:          schema.TypeString,
							Optional:      true,
							DefaultFunc:   schema.EnvDefaultFunc("DNS_UPDATE_KEYSECRET", nil),
							ConflictsWith: []string{"update.0.gssapi.0"},
							RequiredWith:  []string{"update.0.key_name", "update.0.key_algorithm"},
							Description: "Required if `key_name` is set\nA Base64-encoded string containing the " +
								"shared secret to be used for TSIG.",
						},
						"gssapi": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Description: "A `gssapi` block. Only one `gssapi` block may be in the configuration. " +
								"Conflicts with use of `key_name`, `key_algorithm` and `key_secret`.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"realm": {
										Type:        schema.TypeString,
										Required:    true,
										DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_REALM", nil),
										Description: "The Kerberos realm or Active Directory domain.",
									},
									"username": {
										Type:        schema.TypeString,
										Optional:    true,
										DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_USERNAME", nil),
										Description: "The name of the user to authenticate as. If not set the current " +
											"user session will be used.",
									},
									"password": {
										Type:          schema.TypeString,
										Optional:      true,
										DefaultFunc:   schema.EnvDefaultFunc("DNS_UPDATE_PASSWORD", nil),
										ConflictsWith: []string{"update.0.gssapi.0.keytab"},
										RequiredWith:  []string{"update.0.gssapi.0.username"},
										Sensitive:     true,
										Description: "This or `keytab` is required if `username` is set. The matching " +
											"password for `username`.",
									},
									"keytab": {
										Type:          schema.TypeString,
										Optional:      true,
										DefaultFunc:   schema.EnvDefaultFunc("DNS_UPDATE_KEYTAB", nil),
										ConflictsWith: []string{"update.0.gssapi.0.password"},
										RequiredWith:  []string{"update.0.gssapi.0.username"},
										Description: "This or `password` is required if `username` is set, not " +
											"supported on Windows. The path to a keytab file containing a key for " +
											"`username`.",
									},
								},
							},
							ConflictsWith: []string{"update.0.key_name", "update.0.key_algorithm", "update.0.key_secret"},
						},
					},
				},
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"dns_a_record_set":    resourceDnsARecordSet(),
			"dns_aaaa_record_set": resourceDnsAAAARecordSet(),
			"dns_txt_record_set":  resourceDnsTXTRecordSet(),
		},

		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	var server, transport, timeout, keyname, keyalgo, keysecret, realm, username, password, keytab string
	var port, retries int
	var duration time.Duration
	var gssapi bool

	// if the update block is missing, schema.EnvDefaultFunc is not called
	if v, ok := d.GetOk("update"); ok {
		//nolint:forcetypeassert
		update := v.([]interface{})[0].(map[string]interface{})
		if val, ok := update["port"]; ok {
			//nolint:forcetypeassert
			port = val.(int)
		}
		if val, ok := update["server"]; ok {
			//nolint:forcetypeassert
			server = val.(string)
		}
		if val, ok := update["transport"]; ok {
			//nolint:forcetypeassert
			transport = val.(string)
		}
		if val, ok := update["timeout"]; ok {
			//nolint:forcetypeassert
			timeout = val.(string)
		}
		if val, ok := update["retries"]; ok {
			//nolint:forcetypeassert
			retries = val.(int)
		}
		if val, ok := update["key_name"]; ok {
			//nolint:forcetypeassert
			keyname = val.(string)
		}
		if val, ok := update["key_algorithm"]; ok {
			//nolint:forcetypeassert
			keyalgo = val.(string)
		}
		if val, ok := update["key_secret"]; ok {
			//nolint:forcetypeassert
			keysecret = val.(string)
		}
		//nolint:forcetypeassert
		if val, ok := update["gssapi"]; ok && len(val.([]interface{})) > 0 {
			//nolint:forcetypeassert
			g := val.([]interface{})[0].(map[string]interface{})
			if val, ok := g["realm"]; ok {
				//nolint:forcetypeassert
				realm = val.(string)
			}
			if val, ok := g["username"]; ok {
				//nolint:forcetypeassert
				username = val.(string)
			}
			if val, ok := g["password"]; ok {
				//nolint:forcetypeassert
				password = val.(string)
			}
			if val, ok := g["keytab"]; ok {
				//nolint:forcetypeassert
				keytab = val.(string)
			}
			gssapi = true
		}
	} else {
		if len(os.Getenv("DNS_UPDATE_SERVER")) > 0 {
			server = os.Getenv("DNS_UPDATE_SERVER")
		} else {
			return nil, nil
		}
		if len(os.Getenv("DNS_UPDATE_PORT")) > 0 {
			var err error
			portStr := os.Getenv("DNS_UPDATE_PORT")
			port, err = strconv.Atoi(portStr)
			if err != nil {
				return nil, diag.Errorf("invalid DNS_UPDATE_PORT environment variable: %s", err)
			}
		} else {
			port = defaultPort
		}
		if len(os.Getenv("DNS_UPDATE_TRANSPORT")) > 0 {
			transport = os.Getenv("DNS_UPDATE_TRANSPORT")
		} else {
			transport = defaultTransport
		}
		if len(os.Getenv("DNS_UPDATE_TIMEOUT")) > 0 {
			timeout = os.Getenv("DNS_UPDATE_TIMEOUT")
		} else {
			timeout = defaultTimeout
		}
		if len(os.Getenv("DNS_UPDATE_RETRIES")) > 0 {
			var err error
			env := os.Getenv("DNS_UPDATE_RETRIES")
			retries, err = strconv.Atoi(env)
			if err != nil {
				return nil, diag.Errorf(fmt.Sprintf("invalid DNS_UPDATE_RETRIES environment variable: %s", err))
			}
		} else {
			retries = defaultRetries
		}
		if len(os.Getenv("DNS_UPDATE_KEYNAME")) > 0 {
			keyname = os.Getenv("DNS_UPDATE_KEYNAME")
		}
		if len(os.Getenv("DNS_UPDATE_KEYALGORITHM")) > 0 {
			keyalgo = os.Getenv("DNS_UPDATE_KEYALGORITHM")
		}
		if len(os.Getenv("DNS_UPDATE_KEYSECRET")) > 0 {
			keysecret = os.Getenv("DNS_UPDATE_KEYSECRET")
		}
		if len(os.Getenv("DNS_UPDATE_REALM")) > 0 {
			realm = os.Getenv("DNS_UPDATE_REALM")
		}
		if len(os.Getenv("DNS_UPDATE_USERNAME")) > 0 {
			username = os.Getenv("DNS_UPDATE_USERNAME")
		}
		if len(os.Getenv("DNS_UPDATE_PASSWORD")) > 0 {
			password = os.Getenv("DNS_UPDATE_PASSWORD")
		}
		if len(os.Getenv("DNS_UPDATE_KEYTAB")) > 0 {
			keytab = os.Getenv("DNS_UPDATE_KEYTAB")
		}
		if realm != "" || username != "" || password != "" || keytab != "" {
			gssapi = true
		}
	}

	if timeout != "" {
		var err error
		// Try parsing as a duration
		duration, err = time.ParseDuration(timeout)
		if err != nil {
			// Failing that, convert to an integer and treat as seconds
			seconds, err := strconv.Atoi(timeout)
			if err != nil {
				return nil, diag.Errorf("invalid timeout: %s", timeout)
			}
			duration = time.Duration(seconds) * time.Second
		}
		if duration < 0 {
			return nil, diag.Errorf("timeout cannot be negative: %s", duration)
		}
	}

	config := Config{
		server:    server,
		port:      port,
		transport: transport,
		timeout:   duration,
		retries:   retries,
		keyname:   keyname,
		keyalgo:   keyalgo,
		keysecret: keysecret,
		gssapi:    gssapi,
		realm:     realm,
		username:  username,
		password:  password,
		keytab:    keytab,
	}

	return config.Client(ctx)
}

func getAVal(record interface{}) (string, int, error) {

	_, ok := record.(*dns.A)
	if !ok {
		return "", 0, fmt.Errorf("didn't get a A record")
	}

	//nolint:forcetypeassert
	recstr := record.(*dns.A).String()
	var name, class, typ, addr string
	var ttl int

	_, err := fmt.Sscanf(recstr, "%s\t%d\t%s\t%s\t%s", &name, &ttl, &class, &typ, &addr)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing record: %s", err)
	}

	return addr, ttl, nil
}

func getNSVal(record interface{}) (string, int, error) {

	_, ok := record.(*dns.NS)
	if !ok {
		return "", 0, fmt.Errorf("didn't get a NS record")
	}

	//nolint:forcetypeassert
	recstr := record.(*dns.NS).String()
	var name, class, typ, nameserver string
	var ttl int

	_, err := fmt.Sscanf(recstr, "%s\t%d\t%s\t%s\t%s", &name, &ttl, &class, &typ, &nameserver)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing record: %s", err)
	}

	return nameserver, ttl, nil
}

func getAAAAVal(record interface{}) (string, int, error) {

	_, ok := record.(*dns.AAAA)
	if !ok {
		return "", 0, fmt.Errorf("didn't get a AAAA record")
	}

	//nolint:forcetypeassert
	recstr := record.(*dns.AAAA).String()
	var name, class, typ, addr string
	var ttl int

	_, err := fmt.Sscanf(recstr, "%s\t%d\t%s\t%s\t%s", &name, &ttl, &class, &typ, &addr)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing record: %s", err)
	}

	return addr, ttl, nil
}

func getCnameVal(record interface{}) (string, int, error) {

	_, ok := record.(*dns.CNAME)
	if !ok {
		return "", 0, fmt.Errorf("didn't get a CNAME record")
	}

	//nolint:forcetypeassert
	recstr := record.(*dns.CNAME).String()
	var name, class, typ, cname string
	var ttl int

	_, err := fmt.Sscanf(recstr, "%s\t%d\t%s\t%s\t%s", &name, &ttl, &class, &typ, &cname)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing record: %s", err)
	}

	return cname, ttl, nil
}

func getPtrVal(record interface{}) (string, int, error) {

	_, ok := record.(*dns.PTR)
	if !ok {
		return "", 0, fmt.Errorf("didn't get a PTR record")
	}

	//nolint:forcetypeassert
	recstr := record.(*dns.PTR).String()
	var name, class, typ, ptr string
	var ttl int

	_, err := fmt.Sscanf(recstr, "%s\t%d\t%s\t%s\t%s", &name, &ttl, &class, &typ, &ptr)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing record: %s", err)
	}

	return ptr, ttl, nil
}

func isTimeout(err error) bool {

	//nolint:forcetypeassert
	timeout, ok := err.(net.Error)
	return ok && timeout.Timeout()
}

func exchange(msg *dns.Msg, tsig bool, meta interface{}) (*dns.Msg, error) {

	//nolint:forcetypeassert
	c := meta.(*DNSClient).c
	//nolint:forcetypeassert
	srv_addr := meta.(*DNSClient).srv_addr
	//nolint:forcetypeassert
	keyname := meta.(*DNSClient).keyname
	//nolint:forcetypeassert
	keyalgo := meta.(*DNSClient).keyalgo
	//nolint:forcetypeassert
	c.Net = meta.(*DNSClient).transport
	//nolint:forcetypeassert
	retries := meta.(*DNSClient).retries
	//nolint:forcetypeassert
	g := meta.(*DNSClient).gssClient
	retry_tcp := false

	// GSS-TSIG
	if tsig && g != nil {
		//nolint:forcetypeassert
		realm := meta.(*DNSClient).realm
		//nolint:forcetypeassert
		username := meta.(*DNSClient).username
		//nolint:forcetypeassert
		password := meta.(*DNSClient).password
		//nolint:forcetypeassert
		keytab := meta.(*DNSClient).keytab

		var k string
		var err error

		if realm != "" && username != "" && (password != "" || keytab != "") {
			if password != "" {
				k, _, err = g.NegotiateContextWithCredentials(srv_addr, realm, username, password)
			} else {
				k, _, err = g.NegotiateContextWithKeytab(srv_addr, realm, username, keytab)
			}
		} else {
			k, _, err = g.NegotiateContext(srv_addr)
		}
		if err != nil {
			return nil, fmt.Errorf("Error negotiating GSS context: %s", err)
		}

		//nolint:errcheck
		defer g.DeleteContext(k)

		keyname = k
	}

	msg.RecursionDesired = false

	if tsig && keyname != "" {
		msg.SetTsig(keyname, keyalgo, 300, time.Now().Unix())
	}

Retry:
	log.Printf("[DEBUG] Sending DNS message to server (%s):\n%s", srv_addr, msg)

	r, _, err := c.Exchange(msg, srv_addr)

	log.Printf("[DEBUG] Receiving DNS message from server (%s):\n%s", srv_addr, r)

	if err != nil {
		if isTimeout(err) && retries > 0 {
			retries--
			goto Retry
		}
		return r, err
	}

	if r.Rcode == dns.RcodeServerFailure && retries > 0 {
		retries--
		goto Retry
	} else if r.Truncated {
		if retry_tcp {
			switch c.Net {
			case "udp":
				c.Net = "tcp"
			case "udp4":
				c.Net = "tcp4"
			case "udp6":
				c.Net = "tcp6"
			default:
				return nil, fmt.Errorf("Unknown transport: %s", c.Net)
			}
		} else {
			msg.SetEdns0(dns.DefaultMsgSize, false)
			retry_tcp = true
		}

		// Reset retries counter on protocol change
		//nolint:forcetypeassert
		retries = meta.(*DNSClient).retries
		goto Retry
	}

	return r, err
}

func resourceDnsImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	record := d.Id()
	if !dns.IsFqdn(record) {
		return nil, fmt.Errorf("Not a fully-qualified DNS name: %s", record)
	}

	labels := dns.SplitDomainName(record)

	msg := new(dns.Msg)

	var zone *string

Loop:
	for l := range labels {

		msg.SetQuestion(dns.Fqdn(strings.Join(labels[l:], ".")), dns.TypeSOA)

		r, err := exchange(msg, true, meta)
		if err != nil {
			return nil, fmt.Errorf("Error querying DNS record: %s", err)
		}

		switch r.Rcode {
		case dns.RcodeSuccess:

			if len(r.Answer) == 0 {
				continue
			}

			for _, ans := range r.Answer {
				switch t := ans.(type) {
				case *dns.SOA:
					zone = &t.Hdr.Name
				case *dns.CNAME:
					continue Loop
				}
			}

			break Loop
		case dns.RcodeNameError:
			continue
		default:
			return nil, fmt.Errorf("Error querying DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode])
		}
	}

	if zone == nil {
		return nil, fmt.Errorf("No SOA record in authority section in response for %s", record)
	}

	common := dns.CompareDomainName(record, *zone)
	if common == 0 {
		return nil, fmt.Errorf("DNS record %s shares no common labels with zone %s", record, *zone)
	}

	//nolint:errcheck
	d.Set("zone", *zone)
	if name := strings.Join(labels[:len(labels)-common], "."); name != "" {
		//nolint:errcheck
		d.Set("name", name)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceFQDN(d *schema.ResourceData) string {

	//nolint:forcetypeassert
	fqdn := d.Get("zone").(string)

	if name, ok := d.GetOk("name"); ok {
		//nolint:forcetypeassert
		fqdn = fmt.Sprintf("%s.%s", name.(string), fqdn)
	}

	return fqdn
}

func resourceDnsRead(d *schema.ResourceData, meta interface{}, rrType uint16) ([]dns.RR, error) {

	if meta != nil {

		fqdn := resourceFQDN(d)

		msg := new(dns.Msg)
		msg.SetQuestion(fqdn, rrType)

		r, err := exchange(msg, true, meta)
		if err != nil {
			return nil, fmt.Errorf("Error querying DNS record: %s", err)
		}
		switch r.Rcode {
		case dns.RcodeSuccess:
			// NS records are returned slightly differently
			if (rrType == dns.TypeNS && len(r.Ns) > 0) || len(r.Answer) > 0 {
				break
			}
			fallthrough
		case dns.RcodeNameError:
			return nil, nil
		default:
			return nil, fmt.Errorf("Error querying DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode])
		}

		if rrType == dns.TypeNS {
			return r.Ns, nil
		}
		return r.Answer, nil
	} else {
		return nil, fmt.Errorf("update server is not set")
	}
}

func resourceDnsDelete(d *schema.ResourceData, meta interface{}, rrType uint16) error {

	if meta != nil {

		fqdn := resourceFQDN(d)

		//nolint:forcetypeassert
		msg := new(dns.Msg)

		//nolint:forcetypeassert
		msg.SetUpdate(d.Get("zone").(string))

		rrStr := fmt.Sprintf("%s 0 %s", fqdn, dns.TypeToString[rrType])

		rr, err := dns.NewRR(rrStr)
		if err != nil {
			return fmt.Errorf("error reading DNS record (%s): %s", rrStr, err)
		}

		msg.RemoveRRset([]dns.RR{rr})

		r, err := exchange(msg, true, meta)
		if err != nil {
			return fmt.Errorf("Error deleting DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error deleting DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode])
		}

		return nil
	} else {
		return fmt.Errorf("update server is not set")
	}
}
