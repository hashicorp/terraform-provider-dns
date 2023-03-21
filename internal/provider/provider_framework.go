package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miekg/dns"
)

var _ provider.Provider = (*dnsProvider)(nil)

func NewFrameworkProvider() provider.Provider {
	return &dnsProvider{}
}

type dnsProvider struct{}

func (p *dnsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dns"
}

func (p *dnsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var providerConfig providerModel
	var providerUpdateConfig providerUpdateModel
	var providerGssapiConfig providerGssapiModel
	var diags diag.Diagnostics

	var server, transport, timeout, keyname, keyalgo, keysecret, realm, username, password, keytab string
	var port, retries int
	var duration time.Duration
	var gssapi bool

	//TODO change to resp.diags.adderror

	resp.Diagnostics.Append(req.Config.Get(ctx, &providerConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !providerConfig.Update.IsNull() {
		providerConfig.Update.ElementsAs(ctx, &providerUpdateConfig, true)
	}

	if providerUpdateConfig.Server.IsNull() && len(os.Getenv("DNS_UPDATE_SERVER")) > 0 {
		server = os.Getenv("DNS_UPDATE_SERVER")
	}

	if providerUpdateConfig.Port.IsNull() {
		if len(os.Getenv("DNS_UPDATE_PORT")) > 0 {
			portStr := os.Getenv("DNS_UPDATE_PORT")
			envPort, err := strconv.Atoi(portStr)
			if err != nil {
				diags.AddError("invalid DNS_UPDATE_PORT environment variable: ", err.Error()) //TODO: check for the formatting %s
				return
			}
			port = envPort
		} else {
			port = defaultPort
		}
	}

	if providerUpdateConfig.Transport.IsNull() {
		if len(os.Getenv("DNS_UPDATE_TRANSPORT")) > 0 {
			transport = os.Getenv("DNS_UPDATE_TRANSPORT")
		} else {
			transport = defaultTransport
		}
	}

	if providerUpdateConfig.Timeout.IsNull() {
		if len(os.Getenv("DNS_UPDATE_TIMEOUT")) > 0 {
			timeout = os.Getenv("DNS_UPDATE_TIMEOUT")
		} else {
			timeout = defaultTimeout
		}

		// Try parsing as a duration
		var err error
		duration, err = time.ParseDuration(timeout)
		if err != nil {
			// Failing that, convert to an integer and treat as seconds
			var seconds int
			seconds, err = strconv.Atoi(timeout)
			if err != nil {
				diags.AddError("invalid timeout: ", err.Error()) //TODO: check for the formatting %s
				return
			}
			duration = time.Duration(seconds) * time.Second
		}
		if duration < 0 {
			diags.AddError("timeout cannot be negative: ", err.Error()) //TODO: check for the formatting %s
			return
		}

	}

	if providerUpdateConfig.Retries.IsNull() {
		if len(os.Getenv("DNS_UPDATE_RETRIES")) > 0 {
			retriesStr := os.Getenv("DNS_UPDATE_RETRIES")

			var err error
			retries, err = strconv.Atoi(retriesStr)
			if err != nil {
				diags.AddError("invalid DNS_UPDATE_RETRIES environment variable: ", err.Error()) //TODO: check for the formatting %s
				return
			}
		} else {
			retries = defaultRetries
		}
	}
	if providerUpdateConfig.KeyName.IsNull() && len(os.Getenv("DNS_UPDATE_KEYNAME")) > 0 {
		keyname = os.Getenv("DNS_UPDATE_KEYNAME")
	}
	if providerUpdateConfig.KeyAlgorithm.IsNull() && len(os.Getenv("DNS_UPDATE_KEYALGORITHM")) > 0 {
		keyalgo = os.Getenv("DNS_UPDATE_KEYALGORITHM")
	}
	if providerUpdateConfig.KeySecret.IsNull() && len(os.Getenv("DNS_UPDATE_KEYSECRET")) > 0 {
		keysecret = os.Getenv("DNS_UPDATE_KEYSECRET")
	}

	if !providerUpdateConfig.Gssapi.IsNull() {
		providerUpdateConfig.Gssapi.ElementsAs(ctx, &providerGssapiConfig, true)
		gssapi = true
	}
	if providerGssapiConfig.Realm.IsNull() && len(os.Getenv("DNS_UPDATE_REALM")) > 0 {
		realm = os.Getenv("DNS_UPDATE_REALM")
	}
	if providerGssapiConfig.Username.IsNull() && len(os.Getenv("DNS_UPDATE_USERNAME")) > 0 {
		username = os.Getenv("DNS_UPDATE_USERNAME")
	}
	if providerGssapiConfig.Password.IsNull() && len(os.Getenv("DNS_UPDATE_PASSWORD")) > 0 {
		password = os.Getenv("DNS_UPDATE_PASSWORD")
	}
	if providerGssapiConfig.Keytab.IsNull() && len(os.Getenv("DNS_UPDATE_KEYTAB")) > 0 {
		keytab = os.Getenv("DNS_UPDATE_KEYTAB")
	}
	if realm != "" || username != "" || password != "" || keytab != "" {
		gssapi = true
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

	resp.ResourceData, _ = config.Client(ctx)
	resp.DataSourceData, _ = config.Client(ctx)
}

func (p *dnsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDnsARecordSetDataSource,
		NewDnsAAAARecordSetDataSource,
	}
}

func (p *dnsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDnsCNAMERecordSetResource,
	}
}

func (p *dnsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"update": schema.ListNestedBlock{
				Description: "When the provider is used for DNS updates, this block is required.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"server": schema.StringAttribute{
							Required:    true,
							Description: "The hostname or IP address of the DNS server to send updates to.",
						},
						"port": schema.Int64Attribute{
							Optional:    true,
							Description: "The target UDP port on the server where updates are sent to. Defaults to `53`.",
						},
						"transport": schema.StringAttribute{
							Optional: true,
							Description: "Transport to use for DNS queries. Valid values are `udp`, `udp4`, `udp6`, " +
								"`tcp`, `tcp4`, or `tcp6`. Any UDP transport will retry automatically with the " +
								"equivalent TCP transport in the event of a truncated response. Defaults to `udp`.",
						},
						"timeout": schema.StringAttribute{
							Optional: true,
							Description: "Timeout for DNS queries. Valid values are durations expressed as `500ms`, " +
								"etc. or a plain number which is treated as whole seconds.",
						},
						"retries": schema.Int64Attribute{
							Optional:    true,
							Description: "How many times to retry on connection timeout. Defaults to `3`.",
						},
						"key_name": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("gssapi")),
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("key_algorithm"),
									path.MatchRelative().AtParent().AtName("key_secret"),
								),
							},
							//ConflictsWith: []string{"update.0.gssapi.0"},
							//RequiredWith:  []string{"update.0.key_algorithm", "update.0.key_secret"},
							Description: "The name of the TSIG key used to sign the DNS update messages.",
						},
						"key_algorithm": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("gssapi")),
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("key_name"),
									path.MatchRelative().AtParent().AtName("key_secret"),
								),
							},
							//ConflictsWith: []string{"update.0.gssapi.0"},
							//RequiredWith:  []string{"update.0.key_name", "update.0.key_secret"},
							Description: "Required if `key_name` is set. When using TSIG authentication, the " +
								"algorithm to use for HMAC. Valid values are `hmac-md5`, `hmac-sha1`, `hmac-sha256` " +
								"or `hmac-sha512`.",
						},
						"key_secret": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("gssapi")),
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("key_name"),
									path.MatchRelative().AtParent().AtName("key_algorithm"),
								),
							},
							//ConflictsWith: []string{"update.0.gssapi.0"},
							//RequiredWith:  []string{"update.0.key_name", "update.0.key_algorithm"},
							Description: "Required if `key_name` is set\nA Base64-encoded string containing the " +
								"shared secret to be used for TSIG.",
						},
					},
					Blocks: map[string]schema.Block{
						"gssapi": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("key_name"),
									path.MatchRelative().AtParent().AtName("key_algorithm"),
									path.MatchRelative().AtParent().AtName("key_algorithm"),
								),
							},
							Description: "A `gssapi` block. Only one `gssapi` block may be in the configuration. " +
								"Conflicts with use of `key_name`, `key_algorithm` and `key_secret`.",
							//ConflictsWith: []string{"update.0.key_name", "update.0.key_algorithm", "update.0.key_secret"},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"realm": schema.StringAttribute{
										Required:    true,
										Description: "The Kerberos realm or Active Directory domain.",
									},
									"username": schema.StringAttribute{
										Optional: true,
										Description: "The name of the user to authenticate as. If not set the current " +
											"user session will be used.",
									},
									"password": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("keytab")),
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("username")),
										},
										//ConflictsWith: []string{"update.0.gssapi.0.keytab"},
										//RequiredWith:  []string{"update.0.gssapi.0.username"},
										Sensitive: true,
										Description: "This or `keytab` is required if `username` is set. The matching " +
											"password for `username`.",
									},
									"keytab": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("username")),
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password")),
										},
										//ConflictsWith: []string{"update.0.gssapi.0.password"},
										//RequiredWith:  []string{"update.0.gssapi.0.username"},
										Description: "This or `password` is required if `username` is set, not " +
											"supported on Windows. The path to a keytab file containing a key for " +
											"`username`.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type providerModel struct {
	Update types.List `tfsdk:"update"` // providerUpdateModel
}

type providerUpdateModel struct {
	Server       types.String `tfsdk:"server"`
	Port         types.Int64  `tfsdk:"port"`
	Transport    types.String `tfsdk:"transport"`
	Timeout      types.String `tfsdk:"timeout"`
	Retries      types.Int64  `tfsdk:"retries"`
	KeyName      types.String `tfsdk:"key_name"`
	KeyAlgorithm types.String `tfsdk:"key_algorithm"`
	KeySecret    types.String `tfsdk:"key_secret"`
	Gssapi       types.List   `tfsdk:"gssapi"` //providerGssapiModel
}

type providerGssapiModel struct {
	Realm    types.String `tfsdk:"realm"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Keytab   types.String `tfsdk:"keytab"`
}

func exchange_framework(msg *dns.Msg, tsig bool, client *DNSClient) (*dns.Msg, error) {

	c := client.c
	srv_addr := client.srv_addr
	keyname := client.keyname
	keyalgo := client.keyalgo
	c.Net = client.transport
	retries := client.retries
	g := client.gssClient
	retry_tcp := false

	// GSS-TSIG
	if tsig && g != nil {
		realm := client.realm
		username := client.username
		password := client.password
		keytab := client.keytab

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
		retries = client.retries
		goto Retry
	}

	return r, err
}

func resourceDnsImport_framework(id string, client *DNSClient) (dnsConfig, error) {
	var config dnsConfig

	record := id
	if !dns.IsFqdn(record) {
		return config, fmt.Errorf("Not a fully-qualified DNS name: %s", record)
	}

	labels := dns.SplitDomainName(record)

	msg := new(dns.Msg)

	var zone *string

Loop:
	for l := range labels {

		msg.SetQuestion(dns.Fqdn(strings.Join(labels[l:], ".")), dns.TypeSOA)

		r, err := exchange_framework(msg, true, client)
		if err != nil {
			return config, fmt.Errorf("Error querying DNS record: %s", err)
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
			return config, fmt.Errorf("Error querying DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode])
		}
	}

	if zone == nil {
		return config, fmt.Errorf("No SOA record in authority section in response for %s", record)
	}

	common := dns.CompareDomainName(record, *zone)
	if common == 0 {
		return config, fmt.Errorf("DNS record %s shares no common labels with zone %s", record, *zone)
	}

	config.Zone = *zone

	if name := strings.Join(labels[:len(labels)-common], "."); name != "" {
		config.Name = name
	}

	return config, nil
}

func resourceFQDN_framework(config dnsConfig) string {

	fqdn := config.Zone
	fqdn = fmt.Sprintf("%s.%s", config.Name, fqdn)
	return fqdn
}

func resourceDnsRead_framework(config dnsConfig, client *DNSClient, rrType uint16) ([]dns.RR, error) {

	fqdn := resourceFQDN_framework(config)

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, rrType)

	r, err := exchange_framework(msg, true, client)
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
}

func resourceDnsDelete_framework(config dnsConfig, client *DNSClient, rrType uint16) error {

	fqdn := resourceFQDN_framework(config)

	//nolint:forcetypeassert
	msg := new(dns.Msg)

	//nolint:forcetypeassert
	msg.SetUpdate(config.Zone)

	rrStr := fmt.Sprintf("%s 0 %s", fqdn, dns.TypeToString[rrType])

	rr, err := dns.NewRR(rrStr)
	if err != nil {
		return fmt.Errorf("error reading DNS record (%s): %s", rrStr, err)
	}

	msg.RemoveRRset([]dns.RR{rr})

	r, err := exchange_framework(msg, true, client)
	if err != nil {
		return fmt.Errorf("Error deleting DNS record: %s", err)
	}
	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("Error deleting DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode])
	}

	return nil
}

type dnsConfig struct {
	Name string
	Zone string
}
