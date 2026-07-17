package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var productKeywords = map[string]string{
	"ecs":            "ecs",
	"redis":          "redis",
	"云服务器":           "ecs",
	"缓存":             "redis",
	"vpc":            "vpc",
	"网络":             "vpc",
	"rds":            "rds-mysql",
	"mysql":          "rds-mysql",
	"数据库":            "rds-mysql",
	"kafka":          "kafka",
	"cdn":            "cdn",
	"iam":            "iam",
	"kms":            "kms",
	"billing":        "billing",
	"cms":            "cms",
	"告警":             "cms",
	"监控":             "cms",
	"eip":            "eip",
	"clb":            "clb",
	"负载均衡":           "clb",
	"nat":            "nat",
	"vpn":            "vpn",
	"dns":            "dns",
	"tos":            "tos",
	"mongodb":        "mongodb",
	"elasticsearch":  "elasticsearch",
	"vke":            "vke",
	"k8s":            "vke",
	"kubernetes":     "vke",
	"容器":             "vke",
}

var symptomPatterns = []struct {
	pattern *regexp.Regexp
	label   string
}{
	{regexp.MustCompile(`(?i)cpu\s*(占用|使用|高|满|告警|usage|high|spike|100)`), "cpu_high"},
	{regexp.MustCompile(`(?i)内存\s*(占用|使用|高|满|告警|usage|high|spike|100|oom)`), "mem_high"},
	{regexp.MustCompile(`(?i)磁盘\s*(满|不足|告警|usage|high|full|空间)`), "disk_high"},
	{regexp.MustCompile(`(?i)latency|延迟|超时|timeout|慢查询|slow`), "latency"},
	{regexp.MustCompile(`(?i)connection|连接.*失败|连接.*超时|断开|refused`), "connection"},
	{regexp.MustCompile(`(?i)(错误|error|异常|fault|fail|panic|crash)`), "errors"},
	{regexp.MustCompile(`(?i)(重启|restart|重启中)`), "restart"},
	{regexp.MustCompile(`(?i)(扩容|缩容|scale|伸缩)`), "scale"},
	{regexp.MustCompile(`(?i)(安全组|security.?group)`), "security_group"},
}

var resourceIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`i-[a-z0-9]{8,}`),
	regexp.MustCompile(`redis-[a-z0-9]{8,}`),
	regexp.MustCompile(`mysql-[a-z0-9]{8,}`),
	regexp.MustCompile(`vpc-[a-z0-9]{8,}`),
	regexp.MustCompile(`eip-[a-z0-9]{8,}`),
	regexp.MustCompile(`sg-[a-z0-9]{8,}`),
	regexp.MustCompile(`clb-[a-z0-9]{8,}`),
	regexp.MustCompile(`kafka-[a-z0-9]{8,}`),
}

// ParseJSON unmarshals a JSON IncidentPayload and validates product_hint is present.
func ParseJSON(raw []byte) (*IncidentPayload, error) {
	var p IncidentPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("parse incident JSON: %w", err)
	}
	if p.ProductHint == "" {
		return nil, fmt.Errorf("parse incident JSON: product_hint is required")
	}
	if p.Source == "" {
		p.Source = "json"
	}
	p.RawInput = string(raw)
	return &p, nil
}

// ParseNaturalLanguage extracts product, symptom, and resource IDs from free-text input.
func ParseNaturalLanguage(input string) (*IncidentPayload, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("parse natural language: input is empty")
	}

	lower := strings.ToLower(input)

	// Product hint via keyword matching (longest match first).
	var productHint string
	for keyword, product := range productKeywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			// Prefer longer keyword match (more specific).
			if len(keyword) > len(productHint) || productHint == "" {
				productHint = product
			}
		}
	}

	// Symptom extraction.
	var symptoms []string
	for _, sp := range symptomPatterns {
		if sp.pattern.MatchString(input) {
			symptoms = append(symptoms, sp.label)
		}
	}

	// Resource ID extraction.
	seen := map[string]bool{}
	var resourceIDs []string
	for _, rp := range resourceIDPatterns {
		for _, id := range rp.FindAllString(input, -1) {
			if !seen[id] {
				seen[id] = true
				resourceIDs = append(resourceIDs, id)
			}
		}
	}

	return &IncidentPayload{
		ProductHint: productHint,
		Symptom:     strings.Join(symptoms, ", "),
		ResourceIDs: resourceIDs,
		RawInput:    input,
		Source:      "natural_language",
	}, nil
}
