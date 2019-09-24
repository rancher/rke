package openstack

import (
	"fmt"

	"github.com/rancher/rke/templates"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	OpenstackCloudProviderName = "openstack"
	OpenstackConfig            = "OpenstackConfig"
)

type CloudProvider struct {
	Config *v3.OpenstackCloudProvider
	Name   string
}

func GetInstance() *CloudProvider {
	return &CloudProvider{}
}

func (p *CloudProvider) Init(cloudProviderConfig v3.CloudProvider) error {
	if cloudProviderConfig.OpenstackCloudProvider == nil {
		return fmt.Errorf("Openstack Cloud Provider Config is empty")
	}
	p.Name = OpenstackCloudProviderName
	if cloudProviderConfig.Name != "" {
		p.Name = cloudProviderConfig.Name
	}
	p.Config = cloudProviderConfig.OpenstackCloudProvider
	return nil
}

func (p *CloudProvider) GetName() string {
	return p.Name
}

func (p *CloudProvider) GenerateCloudConfigFile() (string, error) {
	// Generate INI style configuration from template https://github.com/go-ini/ini/issues/84
	OpenstackConfig := map[string]v3.OpenstackCloudProvider{
		OpenstackConfig: *p.Config,
	}
	return templates.CompileTemplateFromMap(templates.OpenStackCloudProviderTemplate, OpenstackConfig)
}
