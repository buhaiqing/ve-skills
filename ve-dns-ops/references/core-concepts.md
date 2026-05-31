# Core Concepts — Volcengine DNS (DNS解析)

## What is Volcengine DNS?

Volcengine DNS (DNS解析) is a managed Domain Name System (DNS) service that translates human-readable domain names (like `example.com`) into machine-readable IP addresses (like `192.168.1.1`). It provides:

- **High availability** with distributed nameservers across multiple locations
- **Low latency** resolution with intelligent routing
- **Anti-DDoS protection** against DNS amplification and volumetric attacks
- **Full record type support**: A, AAAA, CNAME, MX, TXT, NS, SRV, CAA

## Key Concepts

### Domains

A **domain** is the root namespace under which DNS records are organized. Each domain in Volcengine DNS represents a DNS zone you control.

| Attribute | Description |
|-----------|-------------|
| Domain ID | Unique identifier (e.g., `d-xxxxxxxx`) |
| Domain Name | Fully Qualified Domain Name (FQDN) like `example.com` |
| Status | `active`, `pending`, `suspended` |
| Record Count | Number of DNS records under this domain |

### Record Sets

A **record set** is a DNS record that maps a name (like `www`) to a value (like `192.168.1.1`) with a specific type and TTL.

| Field | Description | Example |
|-------|-------------|---------|
| RR (Host Record) | Subdomain or `@` for root | `www`, `mail`, `@` |
| Type | Record type | `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`, `CAA` |
| Value | Target value | `192.168.1.1`, `mail.example.com` |
| TTL | Time-to-live in seconds | `600` (10 minutes) |
| Priority | MX/SRV priority | `10` (lower = higher priority) |
| Status | Record status | `active`, `disabled` |

### Record Types

#### A Record (Address Record)
Maps a hostname to an IPv4 address.
```
www.example.com.  600  IN  A  192.168.1.1
```

#### AAAA Record (IPv6 Address Record)
Maps a hostname to an IPv6 address.
```
www.example.com.  600  IN  AAAA  2001:db8::1
```

#### CNAME Record (Canonical Name Record)
Aliases one hostname to another. The target must be another domain name, not an IP address.
```
mail.example.com.  600  IN  CNAME  mail.example.net.
```

> **Important:** CNAME records cannot coexist with other record types at the same name (RFC 1034). CNAME cannot be used at the zone apex (`@`).

#### MX Record (Mail Exchange Record)
Specifies mail servers for the domain, with priority values.
```
example.com.  600  IN  MX  10  mail.example.com.
example.com.  600  IN  MX  20  backup-mail.example.com.
```
Lower priority values are preferred.

#### TXT Record (Text Record)
Holds arbitrary text data. Commonly used for:
- **SPF** (Sender Policy Framework): `v=spf1 include:_spf.example.com ~all`
- **DKIM** (DomainKeys Identified Mail): public key for email signing
- **DMARC** (Domain-based Message Authentication): `v=DMARC1; p=reject; rua=mailto:dmarc@example.com`
- **Domain verification**: proving domain ownership to third-party services

#### NS Record (Name Server Record)
Delegates a subdomain to a different set of name servers.
```
subdomain.example.com.  86400  IN  NS  ns1.external-dns.com.
```

#### SRV Record (Service Record)
Specifies the location (hostname + port) of services.
```
_sip._tcp.example.com.  600  IN  SRV  10 60 5060 sip.example.com.
```
Format: `Priority Weight Port Target`

#### CAA Record (Certification Authority Authorization)
Restricts which Certificate Authorities (CAs) can issue SSL/TLS certificates for the domain.
```
example.com.  600  IN  CAA  0 issue "letsencrypt.org"
example.com.  600  IN  CAA  0 issuewild ";"
```

### TTL (Time-To-Live)

TTL determines how long a DNS resolver caches a record before requesting a fresh copy.

| TTL Value | Use Case |
|-----------|----------|
| 60-300s | Before planned changes; emergency updates |
| 600-3600s | Standard production records |
| 86400s (1 day) | Stable, rarely-changing records |
| 604800s (7 days) | NS records, SOA records |

### DNS Resolution Flow

```
User browser
   ↓
Local DNS resolver (ISP / 8.8.8.8)
   ↓
Root nameserver (.)
   ↓
TLD nameserver (.com)
   ↓
Authoritative nameserver (Volcengine DNS)
   ↓
Returns IP address
```

## Resource Hierarchy

```
Account
  └── Domain (e.g., example.com)
        └── Record Sets (A, CNAME, MX, TXT, ...)
              └── Individual Records
```

## DNS Security

- **Anti-DDoS:** Volcengine DNS provides built-in protection against DNS-level DDoS attacks
- **CAA Records:** Restrict which CAs can issue certificates for your domain
- **DNSSEC:** (when available) Adds cryptographic signatures to DNS records to prevent spoofing
- **Rate Limiting:** API-level rate limiting to prevent abuse

## Related Services

| Service | Relationship | Skill |
|---------|-------------|-------|
| CDN | DNS points to CDN endpoints | `ve-cdn-ops` |
| CLB | DNS resolves to CLB VIPs | `ve-clb-ops` |
| ECS | DNS A records point to ECS IPs | `ve-ecs-ops` |
| SSL | CAA records authorize CA certs | `ve-ssl-ops` |
