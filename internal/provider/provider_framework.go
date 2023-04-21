package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

	var server, transport, timeout, keyname, keyalgo, keysecret, realm, username, password, keytab string
	var port, retries int
	var duration time.Duration
	var gssapi bool
	var configErr error

	providerUpdateConfig := make([]providerUpdateModel, 1)
	providerGssapiConfig := make([]providerGssapiModel, 1)

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

	server = providerUpdateConfig[0].Server.ValueString()
	port = int(providerUpdateConfig[0].Port.ValueInt64())
	transport = providerUpdateConfig[0].Transport.ValueString()
	retries = int(providerUpdateConfig[0].Retries.ValueInt64())
	keyname = providerUpdateConfig[0].KeyName.ValueString()
	keyalgo = providerUpdateConfig[0].KeyAlgorithm.ValueString()
	keysecret = providerUpdateConfig[0].KeySecret.ValueString()

	if providerUpdateConfig[0].Server.IsNull() && len(os.Getenv("DNS_UPDATE_SERVER")) > 0 {
		server = os.Getenv("DNS_UPDATE_SERVER")
	}

	if providerUpdateConfig[0].Port.IsNull() {
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

	if providerUpdateConfig[0].Transport.IsNull() {
		transport = defaultTransport

		if len(os.Getenv("DNS_UPDATE_TRANSPORT")) > 0 {
			transport = os.Getenv("DNS_UPDATE_TRANSPORT")
		}
	}

	if providerUpdateConfig[0].Timeout.IsNull() {
		timeout = defaultTimeout

		if len(os.Getenv("DNS_UPDATE_TIMEOUT")) > 0 {
			timeout = os.Getenv("DNS_UPDATE_TIMEOUT")
		}

	} else {
		timeout = providerUpdateConfig[0].Timeout.String()
	}

	// Try parsing timeout as a duration
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

	if providerUpdateConfig[0].Retries.IsNull() {
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
	if providerUpdateConfig[0].KeyName.IsNull() && len(os.Getenv("DNS_UPDATE_KEYNAME")) > 0 {
		keyname = os.Getenv("DNS_UPDATE_KEYNAME")
	}
	if providerUpdateConfig[0].KeyAlgorithm.IsNull() && len(os.Getenv("DNS_UPDATE_KEYALGORITHM")) > 0 {
		keyalgo = os.Getenv("DNS_UPDATE_KEYALGORITHM")
	}
	if providerUpdateConfig[0].KeySecret.IsNull() && len(os.Getenv("DNS_UPDATE_KEYSECRET")) > 0 {
		keysecret = os.Getenv("DNS_UPDATE_KEYSECRET")
	}

	if !providerUpdateConfig[0].Gssapi.IsNull() {
		resp.Diagnostics.Append(providerUpdateConfig[0].Gssapi.ElementsAs(ctx, &providerGssapiConfig, false)...)

		if resp.Diagnostics.HasError() {
			return
		}
		gssapi = true
	}

	realm = providerGssapiConfig[0].Realm.ValueString()
	username = providerGssapiConfig[0].Username.ValueString()
	password = providerGssapiConfig[0].Password.ValueString()
	keytab = providerGssapiConfig[0].Keytab.ValueString()

	if providerGssapiConfig[0].Realm.IsNull() && len(os.Getenv("DNS_UPDATE_REALM")) > 0 {
		realm = os.Getenv("DNS_UPDATE_REALM")
	}
	if providerGssapiConfig[0].Username.IsNull() && len(os.Getenv("DNS_UPDATE_USERNAME")) > 0 {
		username = os.Getenv("DNS_UPDATE_USERNAME")
	}
	if providerGssapiConfig[0].Password.IsNull() && len(os.Getenv("DNS_UPDATE_PASSWORD")) > 0 {
		password = os.Getenv("DNS_UPDATE_PASSWORD")
	}
	if providerGssapiConfig[0].Keytab.IsNull() && len(os.Getenv("DNS_UPDATE_KEYTAB")) > 0 {
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

	resp.ResourceData, configErr = config.Client(ctx)
	if configErr != nil {
		resp.Diagnostics.AddError("Error initializing DNS Client:", configErr.Error())
	}
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

func (m providerUpdateModel) objectType() types.ObjectType {
	return types.ObjectType{AttrTypes: m.objectAttributeTypes()}
}

func (m providerUpdateModel) objectAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"gssapi": types.ListType{
			ElemType: providerGssapiModel{}.objectType(),
		},
		"key_name":      types.StringType,
		"key_algorithm": types.StringType,
		"key_secret":    types.StringType,
		"port":          types.Int64Type,
		"server":        types.StringType,
		"retries":       types.Int64Type,
		"timeout":       types.StringType,
		"transport":     types.StringType,
	}
}

type providerGssapiModel struct {
	Realm    types.String `tfsdk:"realm"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Keytab   types.String `tfsdk:"keytab"`
}

func (m providerGssapiModel) objectType() types.ObjectType {
	return types.ObjectType{AttrTypes: m.objectAttributeTypes()}
}

func (m providerGssapiModel) objectAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"keytab":   types.StringType,
		"password": types.StringType,
		"realm":    types.StringType,
		"username": types.StringType,
	}
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

		r, err := exchange(msg, true, client)
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

	r, err := exchange(msg, true, client)
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

	r, err := exchange(msg, true, client)
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
