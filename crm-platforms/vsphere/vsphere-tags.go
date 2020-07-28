package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/log"
)

const TagFieldGroup = "group"
const TagFieldDomain = "domain"
const TagFieldIp = "ip"
const TagFieldSubnetName = "subnetname"
const TagFieldCidr = "cidr"
const TagFieldVmName = "vmname"
const TagFieldNetName = "netname"

const TagNotFound = "tag not found"

func (v *VSpherePlatform) GetTagFieldMap(tag string) (map[string]string, error) {
	fieldMap := make(map[string]string)
	ts := strings.Split(tag, ",")
	for _, field := range ts {
		fs := strings.Split(field, "=")
		if len(fs) != 2 {
			return nil, fmt.Errorf("incorrectly formatted tag: %s", tag)
		}
		fieldMap[fs[0]] = fs[1]
	}
	return fieldMap, nil
}

func (v *VSpherePlatform) GetValueForTagField(tag string, fieldName string) (string, error) {
	fm, err := v.GetTagFieldMap(tag)
	if err != nil {
		return "", err
	}
	value, ok := fm[fieldName]
	if !ok {
		return "", fmt.Errorf(TagNotFound)
	}
	return value, nil

}

func (v *VSpherePlatform) GetVmIpTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-vmip"
}

func (v *VSpherePlatform) GetSubnetTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-subnet"
}

func (v *VSpherePlatform) GetVMDomainTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-vmdomain"
}

// GetDomainFromTag get the domain from the tag which is always the last field
func (v *VSpherePlatform) GetDomainFromTag(ctx context.Context, tag string) (string, error) {
	return v.GetValueForTagField(tag, TagFieldDomain)
}

func (v *VSpherePlatform) GetVmIpTag(ctx context.Context, group, vmName, network, ipaddr string) string {
	return TagFieldGroup + "=" + group + "," + TagFieldVmName + "=" + vmName + "," + TagFieldNetName + "=" + network + "," + TagFieldIp + "=" + ipaddr + "," + TagFieldDomain + "=" + string(v.vmProperties.Domain)
}

// ParseVMIpTag returns vmname, network, ipaddr, domain
func (v *VSpherePlatform) ParseVMIpTag(ctx context.Context, tag string) (string, string, string, string, error) {
	fm, err := v.GetTagFieldMap(tag)
	if err != nil {
		return "", "", "", "", err
	}
	vmname, ok := fm[TagFieldVmName]
	if !ok {
		return "", "", "", "", fmt.Errorf("No vmname in vmip tag")
	}
	network, ok := fm[TagFieldNetName]
	if !ok {
		return "", "", "", "", fmt.Errorf("No netname in vmip tag")
	}
	ip, ok := fm[TagFieldIp]
	if !ok {
		return "", "", "", "", fmt.Errorf("No ip in vmip tag")
	}
	domain, ok := fm[TagFieldDomain]
	if !ok {
		return "", "", "", "", fmt.Errorf("No domain in vmip tag")
	}
	return vmname, network, ip, domain, nil
}

func (v *VSpherePlatform) GetSubnetTag(ctx context.Context, group, subnetName, cidr string) string {
	return TagFieldGroup + "=" + group + "," + TagFieldSubnetName + "=" + subnetName + "," + TagFieldCidr + "=" + cidr + "," + TagFieldDomain + "=" + string(v.vmProperties.Domain)
}

// ParseSubnetTag returns subnetName, cidr, domain
func (v *VSpherePlatform) ParseSubnetTag(ctx context.Context, tag string) (string, string, string, error) {
	fm, err := v.GetTagFieldMap(tag)
	if err != nil {
		return "", "", "", err
	}
	subnetName, ok := fm[TagFieldSubnetName]
	if !ok {
		return "", "", "", fmt.Errorf("No subnetname in subnet tag")
	}
	cidr, ok := fm[TagFieldCidr]
	if !ok {
		return "", "", "", fmt.Errorf("No cidr in subnet tag")
	}
	domain, ok := fm[TagFieldDomain]
	if !ok {
		return "", "", "", fmt.Errorf("No domain in subnet tag")
	}
	return subnetName, cidr, domain, nil
}

func (v *VSpherePlatform) GetVmDomainTag(ctx context.Context, group, vmName string) string {
	return TagFieldGroup + "=" + group + "," + TagFieldVmName + "=" + vmName + "," + TagFieldDomain + "=" + string(v.vmProperties.Domain)
}

// ParseVMDomainTag returns vmname, domain
func (v *VSpherePlatform) ParseVMDomainTag(ctx context.Context, tag string) (string, string, error) {
	fm, err := v.GetTagFieldMap(tag)
	if err != nil {
		return "", "", err
	}
	vmName, ok := fm[TagFieldVmName]
	if !ok {
		return "", "", fmt.Errorf("No subnetname in vmdomain tag")
	}
	domain, ok := fm[TagFieldDomain]
	if !ok {
		return "", "", fmt.Errorf("No domain in vmdomain tag")
	}
	return vmName, domain, nil
}

func (v *VSpherePlatform) GetIpsFromTagsForVM(ctx context.Context, vmName string, sd *vmlayer.ServerDetail) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetIpsFromTagsForVM", "vmName", vmName)
	tags, err := v.GetTagsForCategory(ctx, v.GetVmIpTagCategory(ctx), vmlayer.VMDomainAny)
	if err != nil {
		return err
	}
	for _, t := range tags {
		vm, net, ip, _, err := v.ParseVMIpTag(ctx, t.Name)
		if err != nil {
			return err
		}
		if vm != vmName {
			continue
		}

		// see if there is an existing port in the server details and update it
		found := false
		for i, s := range sd.Addresses {
			if s.Network == net {
				log.SpanLog(ctx, log.DebugLevelInfra, "Updated address via tag", "vm", vm, "net", net, "ip", ip)
				sd.Addresses[i].ExternalAddr = ip
				sd.Addresses[i].InternalAddr = ip
				found = true
			}
		}
		if !found {
			sip := vmlayer.ServerIP{
				InternalAddr: ip,
				ExternalAddr: ip,
				Network:      net,
				PortName:     vmlayer.GetPortName(vmName, net),
			}
			sd.Addresses = append(sd.Addresses, sip)
			log.SpanLog(ctx, log.DebugLevelInfra, "Added address via tag", "vm", vm, "net", net, "ip", ip)
		}
	}
	return nil
}

func (v *VSpherePlatform) CreateTag(ctx context.Context, tag *vmlayer.TagOrchestrationParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateTag", "tag", tag)

	out, err := v.TimedGovcCommand(ctx, "govc", "tags.create", "-c", tag.Category, tag.Name)
	if err != nil {
		if strings.Contains(string(out), "ALREADY_EXISTS") {
			log.SpanLog(ctx, log.DebugLevelInfra, "Tag already exists", "tag", tag)
			return nil
		}
		return fmt.Errorf("Error in creating tag: %s - %v", tag.Name, err)
	}
	return nil
}

func (v *VSpherePlatform) DeleteTag(ctx context.Context, tagname string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteTag", "tagname", tagname)

	out, err := v.TimedGovcCommand(ctx, "govc", "tags.rm", tagname)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Tag delete fail", "out", out, "err", err)
		return fmt.Errorf("Error in deleting tag: %s - %v", tagname, err)
	}
	return nil
}

func (v *VSpherePlatform) GetTagsForCategory(ctx context.Context, category string, domainMatch vmlayer.VMDomain) ([]GovcTag, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetTagsForCategory", "category", category, "domainMatch", domainMatch)

	out, err := v.TimedGovcCommand(ctx, "govc", "tags.ls", "-c", category, "-json")

	var tags []GovcTag
	var matchedTags []GovcTag
	err = json.Unmarshal(out, &tags)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetTagsForCategory unmarshal fail", "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal govc subnet tags, %v", err)
		return nil, err
	}
	// due to an intermittent govc bug (or maybe a vsphere bug), sometimes the category id in the tag is a UUID instead
	// of a name so we will update it here before returning to get consistent results
	for _, t := range tags {
		domain, err := v.GetDomainFromTag(ctx, t.Name)
		if err != nil {
			return nil, err
		}
		if domainMatch == vmlayer.VMDomainAny || domain == string(domainMatch) {
			if t.Category != category {
				log.SpanLog(ctx, log.DebugLevelInfra, "Updating category for tag", "orig category", t.Category, "category", category)
				t.Category = category
			}
			matchedTags = append(matchedTags, t)
		}
	}
	return matchedTags, nil
}

func (v *VSpherePlatform) GetTagCategories(ctx context.Context) ([]GovcTagCategory, error) {
	dcName := v.GetDatacenterName(ctx)

	out, err := v.TimedGovcCommand(ctx, "govc", "tags.category.ls", "-json")
	if err != nil {
		return nil, err
	}

	var foundcats []GovcTagCategory
	var returnedcats []GovcTagCategory
	err = json.Unmarshal(out, &foundcats)
	if err != nil {
		return nil, err

	}
	// exclude the ones not in our datacenter
	for _, c := range foundcats {
		if strings.HasPrefix(c.Name, dcName) {
			returnedcats = append(returnedcats, c)
		}
	}
	return returnedcats, err
}

func (v *VSpherePlatform) GetVmNamesFromTags(ctx context.Context, tags []GovcTag) (map[string]string, error) {
	names := make(map[string]string)
	for _, tag := range tags {
		vmname, _, err := v.ParseVMDomainTag(ctx, tag.Name)
		if err != nil {
			return nil, err
		}
		names[vmname] = vmname
	}
	return names, nil
}
