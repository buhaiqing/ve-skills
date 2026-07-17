package agent

import (
	"fmt"
)

var skillMap = map[string]string{
	"ecs":            "ve-ecs-ops",
	"redis":          "ve-redis-ops",
	"vpc":            "ve-vpc-ops",
	"rds-mysql":      "ve-rds-mysql-ops",
	"kafka":          "ve-kafka-ops",
	"cdn":            "ve-cdn-ops",
	"iam":            "ve-iam-ops",
	"kms":            "ve-kms-ops",
	"billing":        "ve-billing-ops",
	"cms":            "ve-cms-ops",
	"eip":            "ve-eip-ops",
	"clb":            "ve-clb-ops",
	"nat":            "ve-nat-ops",
	"vpn":            "ve-vpn-ops",
	"dns":            "ve-dns-ops",
	"tos":            "ve-tos-ops",
	"mongodb":        "ve-mongodb-ops",
	"elasticsearch":  "ve-elasticsearch-ops",
	"vke":            "ve-vke-ops",
}

// Triage maps an IncidentPayload's ProductHint to the corresponding ve-*-ops skill.
// Unknown products default to ve-cms-ops (Rule 5).
func Triage(payload *IncidentPayload) *TriageResult {
	skill, ok := skillMap[payload.ProductHint]
	if !ok {
		skill = "ve-cms-ops"
	}

	return &TriageResult{
		PrimarySkill: skill,
		Confidence:   confidenceFromSkill(skill, ok),
	}
}

func confidenceFromSkill(_ string, found bool) string {
	if found {
		return "high"
	}
	return "low"
}

// BuildDiagnoseCommand returns a ve CLI command string for diagnosing the given skill.
func BuildDiagnoseCommand(skill string, payload *IncidentPayload) string {
	action := diagnoseActionForSkill(skill)
	region := "cn-beijing"
	return fmt.Sprintf("ve %s %s --Region %s", skillToService(skill), action, region)
}

// skillToService strips the "ve-" prefix and "-ops" suffix to get the service name.
func skillToService(skill string) string {
	s := skill
	if len(s) > 3 && s[:3] == "ve-" {
		s = s[3:]
	}
	if len(s) > 4 && s[len(s)-4:] == "-ops" {
		s = s[:len(s)-4]
	}
	return s
}

// diagnoseActionForSkill returns a Describe* action name based on the skill.
func diagnoseActionForSkill(skill string) string {
	switch skill {
	case "ve-ecs-ops":
		return "DescribeInstances"
	case "ve-redis-ops":
		return "DescribeDBInstanceDetail"
	case "ve-vpc-ops":
		return "DescribeVpcs"
	case "ve-rds-mysql-ops":
		return "DescribeDBInstanceDetail"
	case "ve-kafka-ops":
		return "DescribeInstances"
	case "ve-cdn-ops":
		return "DescribeCdnDomainTopUrlVisit"
	case "ve-iam-ops":
		return "GetUser"
	case "ve-kms-ops":
		return "DescribeKey"
	case "ve-billing-ops":
		return "DescribeBillSummaryByProduct"
	case "ve-cms-ops":
		return "DescribeMetricData"
	case "ve-eip-ops":
		return "DescribeEipAddresses"
	case "ve-clb-ops":
		return "DescribeLoadBalancers"
	case "ve-nat-ops":
		return "DescribeNatGateways"
	case "ve-vpn-ops":
		return "DescribeVpnGateways"
	case "ve-dns-ops":
		return "DescribeZones"
	case "ve-tos-ops":
		return "DescribeAccount"
	case "ve-mongodb-ops":
		return "DescribeDBInstances"
	case "ve-elasticsearch-ops":
		return "DescribeInstances"
	case "ve-vke-ops":
		return "ListClusters"
	default:
		return "DescribeMetricData"
	}
}
