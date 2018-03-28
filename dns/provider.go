package dns

import (
	"fmt"
	"os"
	"strconv"
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
						"protocol": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "udp",
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_PROTOCOL", nil),
						},
						"timeout": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("DNS_UPDATE_TIMEOUT", nil),
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

	var server, keyname, keyalgo, keysecret, protocol string
	var port, timeout int

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
		if val, ok := update["protocol"]; ok {
			protocol = val.(string)
		}
		if val, ok := update["timeout"]; ok {
			timeout = int(val.(int))
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
		if len(os.Getenv("DNS_UPDATE_PROTOCOL")) > 0 {
			protocol = os.Getenv("DNS_UPDATE_PROTOCOL")
		}
		if len(os.Getenv("DNS_UPDATE_TIMEOUT")) > 0 {
			var err error
			timeoutStr := os.Getenv("DNS_UPDATE_TIMEOUT")
			timeout, err = strconv.Atoi(timeoutStr)
			if err != nil {
				return nil, fmt.Errorf("invalid DNS_UPDATE_TIMEOUT environment variable: %s", err)
			}
		} else {
			timeout = 0
		}

	}

	config := Config{
		server:    server,
		port:      port,
		keyname:   keyname,
		keyalgo:   keyalgo,
		keysecret: keysecret,
		protocol:  protocol,
		timeout:   timeout,
	}

	return config.Client()
}

func getAVal(record interface{}) (string, error) {

	recstr := record.(*dns.A).String()
	var name, ttl, class, typ, addr string

	_, err := fmt.Sscanf(recstr, "%s\t%s\t%s\t%s\t%s", &name, &ttl, &class, &typ, &addr)
	if err != nil {
		return "", fmt.Errorf("Error parsing record: %s", err)
	}

	return addr, nil
}

func getNSVal(record interface{}) (string, error) {

	recstr := record.(*dns.NS).String()
	var name, ttl, class, typ, nameserver string

	_, err := fmt.Sscanf(recstr, "%s\t%s\t%s\t%s\t%s", &name, &ttl, &class, &typ, &nameserver)
	if err != nil {
		return "", fmt.Errorf("Error parsing record: %s", err)
	}

	return nameserver, nil
}

func getAAAAVal(record interface{}) (string, error) {

	recstr := record.(*dns.AAAA).String()
	var name, ttl, class, typ, addr string

	_, err := fmt.Sscanf(recstr, "%s\t%s\t%s\t%s\t%s", &name, &ttl, &class, &typ, &addr)
	if err != nil {
		return "", fmt.Errorf("Error parsing record: %s", err)
	}

	return addr, nil
}

func getCnameVal(record interface{}) (string, error) {

	recstr := record.(*dns.CNAME).String()
	var name, ttl, class, typ, cname string

	_, err := fmt.Sscanf(recstr, "%s\t%s\t%s\t%s\t%s", &name, &ttl, &class, &typ, &cname)
	if err != nil {
		return "", fmt.Errorf("Error parsing record: %s", err)
	}

	return cname, nil
}

func getPtrVal(record interface{}) (string, error) {

	recstr := record.(*dns.PTR).String()
	var name, ttl, class, typ, ptr string

	_, err := fmt.Sscanf(recstr, "%s\t%s\t%s\t%s\t%s", &name, &ttl, &class, &typ, &ptr)
	if err != nil {
		return "", fmt.Errorf("Error parsing record: %s", err)
	}

	return ptr, nil
}

func exchange(msg *dns.Msg, tsig bool, meta interface{}) (*dns.Msg, error) {

	c := meta.(*DNSClient).c
	srv_addr := meta.(*DNSClient).srv_addr
	keyname := meta.(*DNSClient).keyname
	keyalgo := meta.(*DNSClient).keyalgo

	// If we allow setting the transport default then adjust these
	// Save desired protocol
	old_protocol := c.Net
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

	// Revert UDP fix to use desired protocol
	c.Net = old_protocol

	return r, err
}
