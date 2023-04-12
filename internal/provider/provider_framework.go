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
										Sensitive: true,
										Description: "This or `keytab` is required if `username` is set. The matching " +
											"password for `username`.",
									},
									"keytab": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("username")),
										},
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

func (p *dnsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var providerConfig providerModel
	var providerUpdateConfig providerUpdateModel
	var providerGssapiConfig providerGssapiModel

	var server, transport, timeout, keyname, keyalgo, keysecret, realm, username, password, keytab string
	var port, retries int
	var duration time.Duration
	var gssapi bool

	resp.Diagnostics.Append(req.Config.Get(ctx, &providerConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !providerConfig.Update.IsNull() {
		resp.Diagnostics.Append(providerConfig.Update.ElementsAs(ctx, &providerUpdateConfig, false)...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	if providerUpdateConfig.Server.IsNull() && len(os.Getenv("DNS_UPDATE_SERVER")) > 0 {
		server = os.Getenv("DNS_UPDATE_SERVER")
	}

	if providerUpdateConfig.Port.IsNull() {
		port = defaultPort

		if len(os.Getenv("DNS_UPDATE_PORT")) > 0 {
			portStr := os.Getenv("DNS_UPDATE_PORT")
			envPort, err := strconv.Atoi(portStr)
			if err != nil {
				resp.Diagnostics.AddError("Invalid DNS_UPDATE_PORT environment variable:", err.Error())
				return
			}
			port = envPort
		}
	}

	if providerUpdateConfig.Transport.IsNull() {
		transport = defaultTransport

		if len(os.Getenv("DNS_UPDATE_TRANSPORT")) > 0 {
			transport = os.Getenv("DNS_UPDATE_TRANSPORT")
		}
	}

	if providerUpdateConfig.Timeout.IsNull() {
		timeout = defaultTimeout

		if len(os.Getenv("DNS_UPDATE_TIMEOUT")) > 0 {
			timeout = os.Getenv("DNS_UPDATE_TIMEOUT")
		}

		// Try parsing as a duration
		var err error
		duration, err = time.ParseDuration(timeout)
		if err != nil {
			// Failing that, convert to an integer and treat as seconds
			var seconds int
			seconds, err = strconv.Atoi(timeout)
			if err != nil {
				resp.Diagnostics.AddError("Invalid timeout:",
					fmt.Sprintf("timeout cannot be parsed as an integer: %s", err.Error()))
				return
			}
			duration = time.Duration(seconds) * time.Second
		}
		if duration < 0 {
			resp.Diagnostics.AddError("Invalid timeout:", "timeout cannot be negative.")
			return
		}

	}

	if providerUpdateConfig.Retries.IsNull() {
		retries = defaultRetries

		if len(os.Getenv("DNS_UPDATE_RETRIES")) > 0 {
			retriesStr := os.Getenv("DNS_UPDATE_RETRIES")

			var err error
			retries, err = strconv.Atoi(retriesStr)
			if err != nil {
				resp.Diagnostics.AddError("Invalid DNS_UPDATE_RETRIES environment variable:", err.Error())
				return
			}
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
		resp.Diagnostics.Append(providerUpdateConfig.Gssapi.ElementsAs(ctx, &providerGssapiConfig, false)...)

		if resp.Diagnostics.HasError() {
			return
		}
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
}

func (p *dnsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDnsCNAMERecordResource,
		NewDnsMXRecordSetResource,
		NewDnsNSRecordSetResource,
		NewDnsPTRRecordResource,
		NewDnsSRVRecordSetResource,
		NewDnsTXTRecordSetResource,
	}
}

func (p *dnsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDnsARecordSetDataSource,
		NewDnsAAAARecordSetDataSource,
		NewDnsCNAMERecordSetDataSource,
		NewDnsMXRecordSetDataSource,
		NewDnsNSRecordSetDataSource,
		NewDnsPTRRecordSetDataSource,
		NewDnsSRVRecordSetDataSource,
		NewDnsTXTRecordSetDataSource,
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

//nolint:unparam
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
			return nil, fmt.Errorf("error negotiating GSS context: %s", err)
		}

		//nolint:errcheck
		defer g.DeleteContext(k)

		keyname = k
	}

	msg.RecursionDesired = false

	if tsig && keyname != "" {
		msg.SetTsig(keyname, keyalgo, 300, time.Now().Unix())
	}

	for ok := true; ok; ok = retries > 0 {
		log.Printf("[DEBUG] Sending DNS message to server (%s):\n%s", srv_addr, msg)

		r, _, err := c.Exchange(msg, srv_addr)

		log.Printf("[DEBUG] Receiving DNS message from server (%s):\n%s", srv_addr, r)

		if err != nil {
			if isTimeout(err) && retries > 0 {
				retries--
				continue
			}
			return r, err
		}

		if r.Rcode == dns.RcodeServerFailure && retries > 0 {
			retries--
			continue
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
					return nil, fmt.Errorf("unknown transport: %s", c.Net)
				}
			} else {
				msg.SetEdns0(dns.DefaultMsgSize, false)
				retry_tcp = true
			}

			// Reset retries counter on protocol change
			retries = client.retries
			continue
		}
		return r, err
	}

	//we should never be hitting this line
	return nil, fmt.Errorf("unable to complete DNS exchange")
}

func resourceDnsImport_framework(id string, client *DNSClient) (dnsConfig, diag.Diagnostics) {
	var config dnsConfig
	var diags diag.Diagnostics

	record := id
	if !dns.IsFqdn(record) {
		diags.AddError("Error importing DNS record:",
			fmt.Sprintf("Not a fully-qualified DNS name: %s", record))
		return config, diags
	}

	labels := dns.SplitDomainName(record)

	msg := new(dns.Msg)

	var zone *string

Loop:
	for l := range labels {

		msg.SetQuestion(dns.Fqdn(strings.Join(labels[l:], ".")), dns.TypeSOA)

		r, err := exchange_framework(msg, true, client)
		if err != nil {
			diags.AddError("Error querying DNS record:", err.Error())
			return config, diags
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
			diags.AddError(fmt.Sprintf("Error querying DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
			return config, diags
		}
	}

	if zone == nil {
		diags.AddError("Error querying DNS record:",
			fmt.Sprintf("No SOA record in authority section in response for %s", record))
		return config, diags
	}

	common := dns.CompareDomainName(record, *zone)
	if common == 0 {
		diags.AddError("Error querying DNS record:",
			fmt.Sprintf("DNS record %s shares no common labels with zone %s", record, *zone))
		return config, diags
	}

	config.Zone = *zone

	if name := strings.Join(labels[:len(labels)-common], "."); name != "" {
		config.Name = name
	}

	return config, nil
}

func resourceFQDN_framework(config dnsConfig) string {

	fqdn := config.Zone
	if config.Name != "" {
		fqdn = fmt.Sprintf("%s.%s", config.Name, fqdn)
	}
	return fqdn
}

func resourceDnsRead_framework(config dnsConfig, client *DNSClient, rrType uint16) ([]dns.RR, diag.Diagnostics) {
	var diags diag.Diagnostics
	fqdn := resourceFQDN_framework(config)

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, rrType)

	r, err := exchange_framework(msg, true, client)
	if err != nil {
		diags.AddError("Error querying DNS record:", err.Error())
		return nil, diags
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
		diags.AddError(fmt.Sprintf("Error querying DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
		return nil, diags
	}

	if rrType == dns.TypeNS {
		return r.Ns, nil
	}
	return r.Answer, nil
}

func resourceDnsDelete_framework(config dnsConfig, client *DNSClient, rrType uint16) diag.Diagnostics {
	var diags diag.Diagnostics

	fqdn := resourceFQDN_framework(config)

	msg := new(dns.Msg)

	msg.SetUpdate(config.Zone)

	rrStr := fmt.Sprintf("%s 0 %s", fqdn, dns.TypeToString[rrType])

	rr, err := dns.NewRR(rrStr)
	if err != nil {
		diags.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
		return diags
	}

	msg.RemoveRRset([]dns.RR{rr})

	r, err := exchange_framework(msg, true, client)
	if err != nil {
		diags.AddError("Error deleting DNS record:", err.Error())
		return diags
	}
	if r.Rcode != dns.RcodeSuccess {
		diags.AddError(fmt.Sprintf("Error deleting DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
		return diags
	}

	return nil
}

type dnsConfig struct {
	Name string
	Zone string
}
