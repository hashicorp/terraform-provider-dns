package dns

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/miekg/dns"
)

// Provider returns a schema.Provider for DNS dynamic updates.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"update": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_SERVER", nil),
						},
						"port": &schema.Schema{
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

								return 53, nil
							},
						},
						"key_name": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_KEYNAME", nil),
						},
						"key_algorithm": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_KEYALGORITHM", nil),
						},
						"key_secret": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_KEYSECRET", nil),
						},
					},
				},
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"dns_a_record_set":     dataSourceDnsARecordSet(),
			"dns_aaaa_record_set":  dataSourceDnsAAAARecordSet(),
			"dns_cname_record_set": dataSourceDnsCnameRecordSet(),
			"dns_txt_record_set":   dataSourceDnsTxtRecordSet(),
			"dns_ns_record_set":    dataSourceDnsNSRecordSet(),
			"dns_ptr_record_set":   dataSourceDnsPtrRecordSet(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"dns_a_record_set":    resourceDnsARecordSet(),
			"dns_ns_record_set":   resourceDnsNSRecordSet(),
			"dns_aaaa_record_set": resourceDnsAAAARecordSet(),
			"dns_cname_record":    resourceDnsCnameRecord(),
			"dns_ptr_record":      resourceDnsPtrRecord(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	var server, keyname, keyalgo, keysecret string
	var port int

	// if the update block is missing, schema.EnvDefaultFunc is not called
	if v, ok := d.GetOk("update"); ok {
		update := v.([]interface{})[0].(map[string]interface{})
		if val, ok := update["port"]; ok {
			port = int(val.(int))
		}
		if val, ok := update["server"]; ok {
			server = val.(string)
		}
		if val, ok := update["key_name"]; ok {
			keyname = val.(string)
		}
		if val, ok := update["key_algorithm"]; ok {
			keyalgo = val.(string)
		}
		if val, ok := update["key_secret"]; ok {
			keysecret = val.(string)
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
				return nil, fmt.Errorf("invalid DNS_UPDATE_PORT environment variable: %s", err)
			}
		} else {
			port = 53
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
	}

	config := Config{
		server:    server,
		port:      port,
		keyname:   keyname,
		keyalgo:   keyalgo,
		keysecret: keysecret,
	}

	return config.Client()
}

func getAVal(record interface{}) (string, int, error) {

	_, ok := record.(*dns.A)
	if !ok {
		return "", 0, fmt.Errorf("didn't get a A record")
	}

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

	recstr := record.(*dns.PTR).String()
	var name, class, typ, ptr string
	var ttl int

	_, err := fmt.Sscanf(recstr, "%s\t%d\t%s\t%s\t%s", &name, &ttl, &class, &typ, &ptr)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing record: %s", err)
	}

	return ptr, ttl, nil
}

func exchange(msg *dns.Msg, tsig bool, meta interface{}) (*dns.Msg, error) {

	c := meta.(*DNSClient).c
	srv_addr := meta.(*DNSClient).srv_addr
	keyname := meta.(*DNSClient).keyname
	keyalgo := meta.(*DNSClient).keyalgo

	// If we allow setting the transport default then adjust these
	c.Net = "udp"
	retry_tcp := false

	msg.RecursionDesired = false

Retry:
	if tsig && keyname != "" {
		msg.SetTsig(keyname, keyalgo, 300, time.Now().Unix())
	}

	r, _, err := c.Exchange(msg, srv_addr)

	switch err {
	case dns.ErrTruncated:
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

		goto Retry
	case nil:
		fallthrough
	default:
		// do nothing
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
	for l, _ := range labels {

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
				}
			}

			break Loop
		case dns.RcodeNameError:
			continue
		default:
			return nil, fmt.Errorf("Error querying DNS record: %s", dns.RcodeToString[r.Rcode])
		}
	}

	if zone == nil {
		return nil, fmt.Errorf("No SOA record in authority section in response for %s", record)
	}

	common := dns.CompareDomainName(record, *zone)
	if common == 0 {
		return nil, fmt.Errorf("DNS record %s shares no common labels with zone %s", record, *zone)
	}

	d.Set("zone", *zone)
	if name := strings.Join(labels[:len(labels)-common], "."); name != "" {
		d.Set("name", name)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceFQDN(d *schema.ResourceData) string {

	fqdn := d.Get("zone").(string)

	if name, ok := d.GetOk("name"); ok {
		fqdn = fmt.Sprintf("%s.%s", name.(string), fqdn)
	}

	return fqdn
}
