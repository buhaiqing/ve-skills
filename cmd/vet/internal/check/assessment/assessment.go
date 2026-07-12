// Package assessment validates Worker Output Contract example JSON in
// well-architected-assessment.md files. Faithful Go port of
// scripts/validate_product_assessment.py.
package assessment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var requiredTop = map[string]bool{
	"skill_id": true, "product": true, "region": true, "scope": true,
	"assessment_date": true, "status": true, "partial": true, "resource_count": true,
	"pillars": true, "recommendations": true, "trace": true, "errors": true,
}

var pillars = map[string]bool{"reliability": true, "security": true, "cost": true, "efficiency": true}
var pillarPrefix = map[string]string{"rel": "reliability", "sec": "security", "cost": "cost", "eff": "efficiency"}
var statuses = map[string]bool{"OK": true, "PARTIAL": true, "ERROR": true}
var pillarStatus = map[string]bool{"assessed": true, "not_assessed": true, "skipped": true}
var severities = map[string]bool{"Critical": true, "High": true, "Medium": true, "Low": true}
var confidence = map[string]bool{"HIGH": true, "MEDIUM": true, "LOW": true}
var effort = map[string]bool{"quick": true, "medium": true, "major": true}

var findingIDRe = regexp.MustCompile(`^([a-z0-9]+)-(rel|sec|cost|eff)-(\d{3})$`)
var jsonBlockRe = regexp.MustCompile("(?s)```json\\s*\\n(\\{.*?\\})\\n```")

// ValidateFinding validates a single finding object under pillars.<pillarKey>.
func ValidateFinding(product, pillarKey string, f map[string]any, src string) []string {
	var errs []string
	fid, _ := f["id"].(string)
	m := findingIDRe.FindStringSubmatch(fid)
	if m == nil {
		return append(errs, src+": finding id '"+fid+"' invalid (expected {product}-{rel|sec|cost|eff}-NNN)")
	}
	if m[1] != product {
		errs = append(errs, src+": finding id product prefix '"+m[1]+"' != top-level product '"+product+"'")
	}
	if pillarPrefix[m[2]] != pillarKey {
		errs = append(errs, src+": finding '"+fid+"' in pillars."+pillarKey+" but id implies pillars."+pillarPrefix[m[2]])
	}
	for _, field := range []string{"severity", "confidence", "title", "evidence", "recommendation", "effort"} {
		if _, ok := f[field]; !ok {
			errs = append(errs, src+": finding '"+fid+"' missing '"+field+"'")
		}
	}
	sev, _ := f["severity"].(string)
	if !severities[sev] {
		errs = append(errs, src+": finding '"+fid+"' bad severity")
	}
	conf, _ := f["confidence"].(string)
	if !confidence[conf] {
		errs = append(errs, src+": finding '"+fid+"' bad confidence")
	}
	eff, _ := f["effort"].(string)
	if !effort[eff] {
		errs = append(errs, src+": finding '"+fid+"' bad effort")
	}
	return errs
}

// ValidateAssessment validates a single parsed product_assessment object.
func ValidateAssessment(data any, source string) []string {
	var errs []string
	obj, ok := data.(map[string]any)
	if !ok {
		return append(errs, source+": not a JSON object")
	}

	missing := []string{}
	for k := range requiredTop {
		if _, present := obj[k]; !present {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		errs = append(errs, source+": missing top-level fields: "+strings.Join(missing, " "))
	}
	status, _ := obj["status"].(string)
	if !statuses[status] {
		errs = append(errs, source+": invalid status '"+fmtVal(obj["status"])+"'")
	}
	product, _ := obj["product"].(string)
	if product == "" {
		errs = append(errs, source+": product must be non-empty string")
	}

	pv, ok := obj["pillars"].(map[string]any)
	if !ok {
		errs = append(errs, source+": pillars must be object")
		return errs
	}
	for pk, pval := range pv {
		if !pillars[pk] {
			errs = append(errs, source+": unknown pillar key '"+pk+"'")
			continue
		}
		pm, ok := pval.(map[string]any)
		if !ok {
			errs = append(errs, source+": pillars."+pk+" must be object")
			continue
		}
		pst, _ := pm["status"].(string)
		if !pillarStatus[pst] {
			errs = append(errs, source+": pillars."+pk+".status invalid '"+fmtVal(pm["status"])+"'")
		}
		findings, _ := pm["findings"].([]any)
		for fi, fv := range findings {
			fm, ok := fv.(map[string]any)
			if !ok {
				continue
			}
			if product != "" {
				errs = append(errs, ValidateFinding(product, pk, fm, source+" pillars."+pk+"["+strconv.Itoa(fi)+"]")...)
			}
		}
	}

	if recs, ok := obj["recommendations"].([]any); ok {
		for i, rv := range recs {
			rm, ok := rv.(map[string]any)
			if !ok {
				errs = append(errs, source+": recommendations["+strconv.Itoa(i)+"] not object")
				continue
			}
			pil, _ := rm["pillar"].(string)
			if !pillars[pil] {
				errs = append(errs, source+": recommendations["+strconv.Itoa(i)+"].pillar invalid")
			}
		}
	}

	if trace, ok := obj["trace"].(map[string]any); ok {
		if cmds, ok := trace["commands"].([]any); ok {
			for _, cv := range cmds {
				c, _ := cv.(string)
				if strings.Contains(c, "SecretKey=") && !strings.Contains(c, "<masked>") {
					errs = append(errs, source+": trace.commands contains unmasked SecretKey")
				}
				if strings.Contains(c, "VOLCENGINE_SECRET_KEY=") && !strings.Contains(c, "<masked>") {
					errs = append(errs, source+": trace.commands contains unmasked VOLCENGINE_SECRET_KEY")
				}
				if strings.Contains(c, "AKLT") && !strings.Contains(c, "<masked>") {
					errs = append(errs, source+": trace.commands contains unmasked AKLT token")
				}
			}
		}
	}
	return errs
}

// CheckDir scans well-architected-assessment.md files under ve-*-ops and
// returns a per-file error map (only files with errors are present) plus the
// sorted file list — matching the frontmatter package's CheckDir signature.
func CheckDir(root string) (map[string][]string, []string) {
	results := make(map[string][]string)
	pattern := filepath.Join(root, "ve-*-ops", "references", "well-architected-assessment.md")
	matches, _ := filepath.Glob(pattern)
	sort.Strings(matches)

	for _, md := range matches {
		text, err := os.ReadFile(md)
		if err != nil {
			continue
		}
		t := string(text)
		var fileErrs []string
		if !strings.Contains(t, "Worker Output Contract") {
			fileErrs = append(fileErrs, md+": missing 'Worker Output Contract' section")
		}
		examples := extractExamples(t)
		if len(examples) == 0 {
			fileErrs = append(fileErrs, md+": no product_assessment JSON example found")
			results[md] = fileErrs
			continue
		}
		line := 1
		for _, raw := range examples {
			var obj any
			if err := json.Unmarshal([]byte(raw), &obj); err != nil {
				fileErrs = append(fileErrs, md+": JSON parse error: "+err.Error())
				continue
			}
			fileErrs = append(fileErrs, ValidateAssessment(obj, md+":"+strconv.Itoa(line))...)
		}
		if len(fileErrs) > 0 {
			results[md] = fileErrs
		}
	}
	return results, matches
}

func extractExamples(text string) []string {
	var out []string
	for _, m := range jsonBlockRe.FindAllStringSubmatch(text, -1) {
		raw := m[1]
		if !strings.Contains(raw, `"product"`) || !strings.Contains(raw, `"pillars"`) {
			continue
		}
		out = append(out, raw)
	}
	return out
}

func fmtVal(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return "nil"
	default:
		return ""
	}
}
