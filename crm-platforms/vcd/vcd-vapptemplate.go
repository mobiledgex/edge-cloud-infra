package vcd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

type ArtifactoryTokenResp struct {
	Scope       string `json:"scope"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// Return requested vdc template
func (v *VcdPlatform) FindTemplate(ctx context.Context, tmplName string, vcdClient *govcd.VCDClient) (*govcd.VAppTemplate, error) {

	log.SpanLog(ctx, log.DebugLevelInfra, "Find template", "Name", tmplName)
	tmpls, err := v.GetAllVdcTemplates(ctx, vcdClient)
	if err != nil {
		return nil, err
	}

	for _, tmpl := range tmpls {
		if tmpl.VAppTemplate.Name == tmplName {
			log.SpanLog(ctx, log.DebugLevelInfra, "Found template", "Name", tmplName)
			return tmpl, nil
		}
	}

	return nil, fmt.Errorf("template %s not found", tmplName)
}

var vcdDirect string = "vcdDirect"

func (v *VcdPlatform) GetArtifactoryToken(ctx context.Context, host string) (string, error) {
	log.WarnLog("XXX GetArtifactoryToken", "host", host)

	url := fmt.Sprintf("https://%s/artifactory/api/security/token", host)
	reqConfig := cloudcommon.RequestConfig{}
	reqConfig.Headers = make(map[string]string)
	reqConfig.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	resp, err := cloudcommon.SendHTTPReq(ctx, "POST", url, v.vmProperties.CommonPf.PlatformConfig.AccessApi, &reqConfig, strings.NewReader("username="+vcdDirect+"&scope=member-of-groups:readers"))
	log.WarnLog("XXX GetArtifactoryToken", "err", err)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("error reading gettoken response: %v", err)
	}
	var tokResp ArtifactoryTokenResp
	err = json.Unmarshal(body, &tokResp)
	if err != nil {
		return "", fmt.Errorf("Fail to unmarshal response - %v", err)
	}
	log.InfoLog("XXX GetArtifactoryToken got token", "token", tokResp.AccessToken)
	return tokResp.AccessToken, nil
}

func (v *VcdPlatform) ImportTemplateFromUrl(ctx context.Context, name, templUrl string, catalog *govcd.Catalog) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ImportTemplateFromUrl", "name", name, "templUrl", templUrl)
	err := catalog.UploadOvfUrl(templUrl, name, name)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed UploadOvfUrl", "err", err)
		return fmt.Errorf("Failed to upload from URL - %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "ImportTemplateFromUrl done")
	return nil
}

// Return all templates found as vdc resources from MEX_CATALOG
func (v *VcdPlatform) GetAllVdcTemplates(ctx context.Context, vcdClient *govcd.VCDClient) ([]*govcd.VAppTemplate, error) {

	var tmpls []*govcd.VAppTemplate
	org, err := v.GetOrg(ctx, vcdClient)
	if err != nil {
		return tmpls, err
	}
	vdc, err := v.GetVdc(ctx, vcdClient)
	if err != nil {
		return tmpls, err
	}
	// Get our catalog MEX_CATALOG
	catName := v.GetCatalogName()
	if catName == "" {
		return tmpls, fmt.Errorf("MEX_CATALOG name not found")
	}

	cat, err := org.GetCatalogByName(catName, true)
	if err != nil {
		return tmpls, err
	}

	for _, r := range vdc.Vdc.ResourceEntities {
		for _, res := range r.ResourceEntity {
			if res.Type == "application/vnd.vmware.vcloud.vAppTemplate+xml" {
				if v.Verbose {
					log.SpanLog(ctx, log.DebugLevelInfra, "Found Vdc resource template", "Name", res.Name, "from Catalog", catName)
				}
				tmpl, err := cat.GetVappTemplateByHref(res.HREF)
				if err != nil {
					continue
				} else {
					tmpls = append(tmpls, tmpl)
				}
			}
		}
	}
	return tmpls, nil
}
