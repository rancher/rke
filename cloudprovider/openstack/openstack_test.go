package openstack

import (
	"testing"

	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const osCloudINI = `[Global]
auth-url = "https://someurl"
username = "myuser"
user-id = "myuserid"
password = "mypassword"
tenant-id = "mytenantid"
tenant-name = "mytenant"
trust-id = "mytrustid"
domain-id = "mydomainid"
domain-name = "mydomain"
region = "myregion"
ca-file = "mycafile"

[LoadBalancer]
lb-version = "v2"
use-octavia = true
subnet-id = "fa6a4e6c-6ae4-4dde-ae86-3e2f452c1f03"
floating-network-id = "a57af0a0-da92-49be-a98a-345ceca004b3"
lb-method = "method"
lb-provider = "default"
create-monitor = true
monitor-delay = "60s"
monitor-timeout = "30s"
monitor-max-retries = 5
manage-security-groups = true

[BlockStorage]
bs-version = "v2"
trust-device-path = true
ignore-volume-az = true

[Route]
router-id = "e0bf297b-b058-4b41-a706-0b8d21fad43e"

[Metadata]
search-order = "asc"
request-timeout = 10
`

/*
floating-subnet-id = "a02eb6c3-fc69-46ae-a3fe-fb43c1563cbc"

[LoadBalancerClass "dmz"]
floating-network-id = "a374bed4-e920-4c40-b646-2d8927f7f67b"

[LoadBalancerClass "internetFacing"]
floating-network-id = "c57af0a0-da92-49be-a98a-345ceca004b3"
floating-subnet-id = "f90d2440-d3c6-417a-a696-04e55eeb9279"
*/

func TestOpenStackGenerateCloudConfigFile(t *testing.T) {
	cp := CloudProvider{
		Config: &v3.OpenstackCloudProvider{
			Global: v3.GlobalOpenstackOpts{
				AuthURL:    "https://someurl",
				Username:   "myuser",
				UserID:     "myuserid",
				Password:   "mypassword",
				TenantID:   "mytenantid",
				TenantName: "mytenant",
				TrustID:    "mytrustid",
				DomainID:   "mydomainid",
				DomainName: "mydomain",
				Region:     "myregion",
				CAFile:     "mycafile",
			},
			LoadBalancer: v3.LoadBalancerOpenstackOpts{
				LBVersion:         "v2",
				UseOctavia:        true,
				SubnetID:          "fa6a4e6c-6ae4-4dde-ae86-3e2f452c1f03",
				FloatingNetworkID: "a57af0a0-da92-49be-a98a-345ceca004b3",
				// FloatingSubnetID:     "a02eb6c3-fc69-46ae-a3fe-fb43c1563cbc",
				LBMethod:             "method",
				LBProvider:           "default",
				CreateMonitor:        true,
				MonitorDelay:         "60s",
				MonitorTimeout:       "30s",
				MonitorMaxRetries:    5,
				ManageSecurityGroups: true,
				/*
					LBClasses: map[string]v3.LoadBalancerClassOpenstackOpts{
						"internetFacing": {
							FloatingNetworkID: "c57af0a0-da92-49be-a98a-345ceca004b3",
							FloatingSubnetID:  "f90d2440-d3c6-417a-a696-04e55eeb9279",
						},
						"dmz": {
							FloatingNetworkID: "a374bed4-e920-4c40-b646-2d8927f7f67b",
						},
					},
				*/
			},
			BlockStorage: v3.BlockStorageOpenstackOpts{
				BSVersion:       "v2",
				TrustDevicePath: true,
				IgnoreVolumeAZ:  true,
			},
			Route: v3.RouteOpenstackOpts{
				RouterID: "e0bf297b-b058-4b41-a706-0b8d21fad43e",
			},
			Metadata: v3.MetadataOpenstackOpts{
				SearchOrder:    "asc",
				RequestTimeout: 10,
			},
		},
	}

	generatedINI, err := cp.GenerateCloudConfigFile()
	if err != nil {
		t.Fatal(err)
	}

	if generatedINI != osCloudINI {
		t.Fatalf("Generated OpenStack INI file doesn't correspond to expected content:\n%s", generatedINI)
	}
}
