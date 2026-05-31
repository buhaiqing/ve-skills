# DNS Knowledge Base

## Pattern: DNS Resolution Failure

**Root Causes:**
1. Domain not registered or expired
2. Record set not created or incorrect
3. TTL propagation delay
4. DNS server not responding

**Resolution Steps:**
1. Verify domain registration
2. Check record set configuration
3. Test with `dig` or `nslookup`
4. Wait for TTL expiry

## Pattern: Slow DNS Propagation

**Root Causes:**
1. Long TTL values
2. ISP DNS caching
3. Global propagation delay

**Resolution Steps:**
1. Reduce TTL before planned changes
2. Flush local DNS cache
3. Verify with global DNS checker tools
