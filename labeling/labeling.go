package labeling

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/marthjod/gocart/ocatypes"
	"github.com/marthjod/gocart/vmpool"
	"github.com/marthjod/opennebula-exporter/config"
)

func AddLabels(cfg config.Config, vmPool *vmpool.VmPool) string {
	var lines strings.Builder

	var apiHost string
	apiURL, err := url.Parse(cfg.API.Endpoint)
	if err != nil {
		apiHost = cfg.API.Endpoint
	} else {
		apiHost = apiURL.Hostname()
	}

	for _, vm := range vmPool.Vms {
		var b strings.Builder
		fmt.Fprintf(&b, `%s_vms{name=%q,id="%d",lcm_state=%q,api_host=%q`,
			cfg.Exporter.Namespace, vm.Name, vm.Id, vm.LCMState, apiHost)

		if len(cfg.VMNameRegexpLabels) > 0 {
			b.WriteString(AddVMNameRegexpLabels(vm, cfg.VMNameRegexpLabels))
		}

		if len(cfg.UserTemplateLabels) > 0 {
			b.WriteString(AddUserTemplateLabels(vm, cfg.UserTemplateLabels))
		}

		b.WriteString("} 1\n")
		lines.WriteString(b.String())
	}

	return lines.String()
}

func AddUserTemplateLabels(vm *ocatypes.Vm, labels []config.UserTemplateLabel) string {
	var labelAttrs []string

	for _, label := range labels {
		field, err := vm.UserTemplate.Items.GetCustom(label.TemplateField)
		if err != nil {
			field = "unknown"
		}
		labelAttrs = append(labelAttrs, fmt.Sprintf(`%s=%q`, label.Name, field))
	}

	return buildString(labelAttrs)

}

// TODO compile label regexps only once
func AddVMNameRegexpLabels(vm *ocatypes.Vm, labels []config.VMNameRegexpLabel) string {
	var labelAttrs []string

	for _, label := range labels {
		labelMatch, err := regexp.Compile(label.Regexp)
		if err != nil {
			labelAttrs = append(labelAttrs, fmt.Sprintf(`%s=%q`, label.Name, err))
			continue
		}

		if labelMatch.MatchString(vm.Name) {
			matches := labelMatch.FindStringSubmatch(vm.Name)
			// not checking for nil here since it matched before
			match := matches[len(matches)-1]
			labelAttrs = append(labelAttrs, fmt.Sprintf(`%s=%q`, label.Name, match))
		}
	}

	return buildString(labelAttrs)
}

func buildString(a []string) string {
	if len(a) > 0 {
		var b strings.Builder
		b.WriteString(",")
		b.WriteString(strings.Join(a, ","))
		return b.String()
	}

	return ""
}
