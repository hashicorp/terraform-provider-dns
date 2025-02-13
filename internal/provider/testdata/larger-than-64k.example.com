$TTL 86400
@		IN	SOA	ns.example.com. hostmaster.example.com. (
				2021011301 ; serial
				60         ; refresh (1 minute)
				15         ; retry (15 seconds)
				1800       ; expire (30 minutes)
				10         ; minimum (10 seconds)
				)
		IN	NS	ns.example.com.
ns		IN	A	127.0.0.1
