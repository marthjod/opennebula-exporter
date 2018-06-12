package labeling

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/marthjod/gocart/ocatypes"
	"github.com/marthjod/gocart/vmpool"
	"github.com/marthjod/opennebula-exporter/config"
)

type regexpLabel struct {
	Name   string
	Regexp *regexp.Regexp
}

func AddLabels(cfg config.Config, vmPool *vmpool.VmPool) string {
	var (
		lines        strings.Builder
		regexpLabels = []*regexpLabel{}
	)

	if len(cfg.VMNameRegexpLabels) > 0 {
		regexpLabels = compileRegexpLabels(cfg.VMNameRegexpLabels)
	}

	for _, vm := range vmPool.Vms {
		var b strings.Builder
		fmt.Fprintf(&b, `%s_vms{name=%q,id="%d",lcm_state=%q`,
			cfg.Exporter.Namespace, vm.Name, vm.Id, vm.LCMState)

		// even if regexpLabels is empty, check length to avoid func call
		if len(cfg.VMNameRegexpLabels) > 0 {
			b.WriteString(AddVMNameRegexpLabels(vm, regexpLabels))
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

func AddVMNameRegexpLabels(vm *ocatypes.Vm, labels []*regexpLabel) string {
	var labelAttrs []string

	for _, label := range labels {
		if label.Regexp.MatchString(vm.Name) {
			matches := label.Regexp.FindStringSubmatch(vm.Name)
			// not checking for nil here since it matched before
			match := matches[len(matches)-1]
			labelAttrs = append(labelAttrs, fmt.Sprintf(`%s=%q`, label.Name, match))
		}
	}

	return buildString(labelAttrs)
}

func compileRegexpLabels(expressions []config.VMNameRegexpLabel) []*regexpLabel {
	var regexpLabels = []*regexpLabel{}

	for _, label := range expressions {
		re, err := regexp.Compile(label.Regexp)
		if err != nil {
			fmt.Printf("# error compiling regexp for name %q\n", label.Name)
			continue
		}

		regexpLabels = append(regexpLabels, &regexpLabel{
			Name:   label.Name,
			Regexp: re,
		})
	}

	return regexpLabels
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
