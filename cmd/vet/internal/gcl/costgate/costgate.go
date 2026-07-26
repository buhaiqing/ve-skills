package costgate

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type OperationCost struct {
	Operation     string
	BillingModel  string
	ResourceType  string
	EstMonthlyCost float64
	RefundOnDelete float64
	NewMonthlyCost float64
	NetMonthlyDelta float64
	Warning       string
}

var (
	destructiveVerbs = regexp.MustCompile(`\b(Delete|Stop|Terminate|Release|Remove|Shutdown|PowerOff)`)
	createVerbs      = regexp.MustCompile(`\b(Create|Allocate|Provision|Add)`)
	scaleVerbs       = regexp.MustCompile(`\b(Scale|Modify|Resize|Upgrade|Convert|Attach|Detach)`)
	instanceIDRe     = regexp.MustCompile(`(?:--InstanceId|--DBInstanceId|--InstanceID|--ResourceId)\s+[\"']?(\S+)[\"']?`)
	instanceTypeRe   = regexp.MustCompile(`(?:--InstanceType|--InstanceSpec|--InstanceClass)\s+(\S+)`)
)

const warningThreshold = 100.0

func EstimateCostImpact(skill, command string) *OperationCost {
	if command == "" {
		return nil
	}
	product := strings.TrimPrefix(strings.TrimSuffix(skill, "-ops"), "ve-")
	op := extractOperation(command)
	if op == "unknown" {
		return nil
	}

	cost := &OperationCost{
		Operation:    op,
		ResourceType: product,
	}

	switch {
	case destructiveVerbs.MatchString(command):
		cost.BillingModel = queryBillingModel(command)
		cost.EstMonthlyCost = queryCurrentCost(command)
		cost.RefundOnDelete = calculateRefund(cost.BillingModel, cost.EstMonthlyCost)
		cost.NewMonthlyCost = 0
		cost.NetMonthlyDelta = -cost.RefundOnDelete
		if cost.BillingModel == "Spot" {
			cost.Warning = "Spot instance — may be interrupted at any time"
		} else if cost.RefundOnDelete > warningThreshold {
			cost.Warning = "Refund > warning threshold — confirm resource is safe to delete"
		}

	case createVerbs.MatchString(command), scaleVerbs.MatchString(command):
		cost.NewMonthlyCost = queryNewCost(skill, command)
		cost.EstMonthlyCost = 0
		cost.NetMonthlyDelta = cost.NewMonthlyCost
		if cost.NetMonthlyDelta > warningThreshold {
			cost.Warning = "Cost increase > " + strconv.FormatFloat(warningThreshold, 'f', 0, 64) + " CNY/month"
		}
	}

	return cost
}

func extractOperation(cmd string) string {
	switch {
	case destructiveVerbs.MatchString(cmd):
		return "destructive"
	case createVerbs.MatchString(cmd):
		return "create"
	case scaleVerbs.MatchString(cmd):
		return "scale"
	}
	return "unknown"
}

func queryBillingModel(command string) string {
	id := extractResourceID(command)
	if id == "" {
		return "PostPaid"
	}
	// Use sh -c heredoc-style to avoid quoting issues; fallback to PostPaid on error.
	out, err := exec.Command("sh", "-c",
		"ve billing DescribeBillDetail --ResourceId "+id+" 2>/dev/null | jq -r '.Result.BillingModel // empty'").Output()
	if err != nil {
		return "PostPaid"
	}
	model := strings.TrimSpace(string(out))
	if model == "" || model == "empty" {
		return "PostPaid"
	}
	return model
}

func queryCurrentCost(command string) float64 {
	id := extractResourceID(command)
	if id == "" {
		return 0
	}
	out, err := exec.Command("sh", "-c",
		"ve billing DescribeBillDetail --ResourceId "+id+" 2>/dev/null | jq -r '.Result.MonthlyCost | tonumber // 0'").Output()
	if err != nil {
		return 0
	}
	cost, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return cost
}

func queryNewCost(skill, command string) float64 {
	id := extractResourceID(command)
	instanceType := extractInstanceType(command)
	if id == "" && instanceType == "" {
		return 0
	}
	if instanceType != "" {
		return queryPricePerMonth(skill, instanceType)
	}
	// Fallback: query current cost as estimate
	return queryCurrentCost(command)
}

func queryPricePerMonth(skill, instanceType string) float64 {
	out, err := exec.Command("sh", "-c",
		"ve billing QueryPrice --ResourceType "+skill+" --InstanceType "+instanceType+" 2>/dev/null | jq -r '.Result.PricePerMonth | tonumber // 0'").Output()
	if err != nil {
		return 0
	}
	price, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return price
}

func calculateRefund(billingModel string, currentCost float64) float64 {
	switch billingModel {
	case "PrePaid", "Subscription":
		// Rough estimate: unused months refund (simplified)
		return currentCost * 0.8
	case "Spot":
		return 0 // Spot already billed by second
	case "PostPaid":
		return 0 // No refund for pay-as-you-go
	}
	return 0
}

func extractResourceID(command string) string {
	m := instanceIDRe.FindStringSubmatch(command)
	if m == nil {
		return ""
	}
	return m[1]
}

func extractInstanceType(command string) string {
	m := instanceTypeRe.FindStringSubmatch(command)
	if m == nil {
		return ""
	}
	return m[1]
}
