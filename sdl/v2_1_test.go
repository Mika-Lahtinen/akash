package sdl

import (
	"testing"

	manifest "github.com/akash-network/akash-api/go/manifest/v2beta2"
	dtypes "github.com/akash-network/akash-api/go/node/deployment/v1beta3"
	"github.com/akash-network/akash-api/go/node/types/unit"
	atypes "github.com/akash-network/akash-api/go/node/types/v1beta3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "github.com/akash-network/node/testutil"
)

func TestV2_1_ParseSimpleGPU(t *testing.T) {
	sdl, err := ReadFile("./_testdata/v2.1-simple-gpu.yaml")
	require.NoError(t, err)

	groups, err := sdl.DeploymentGroups()
	require.NoError(t, err)
	assert.Len(t, groups, 1)

	group := groups[0]
	assert.Len(t, group.GetResourceUnits(), 1)
	assert.Len(t, group.Requirements.Attributes, 2)

	assert.Equal(t, atypes.Attribute{
		Key:   "region",
		Value: "us-west",
	}, group.Requirements.Attributes[1])

	assert.Len(t, group.GetResourceUnits(), 1)

	assert.Equal(t, dtypes.ResourceUnit{
		Count: 2,
		Resources: atypes.Resources{
			ID: 1,
			CPU: &atypes.CPU{
				Units: atypes.NewResourceValue(randCPU),
			},
			GPU: &atypes.GPU{
				Units: atypes.NewResourceValue(randGPU),
				Attributes: atypes.Attributes{
					{
						Key:   "vendor/nvidia/model/a100",
						Value: "true",
					},
				},
			},
			Memory: &atypes.Memory{
				Quantity: atypes.NewResourceValue(randMemory),
			},
			Storage: atypes.Volumes{
				{
					Name:     "default",
					Quantity: atypes.NewResourceValue(randStorage),
				},
			},
			Endpoints: []atypes.Endpoint{
				{
					Kind: atypes.Endpoint_SHARED_HTTP,
				},
				{
					Kind: atypes.Endpoint_RANDOM_PORT,
				},
			},
		},
		Price: AkashDecCoin(t, 50),
	}, group.GetResourceUnits()[0])

	mani, err := sdl.Manifest()
	require.NoError(t, err)

	assert.Len(t, mani.GetGroups(), 1)

	expectedHosts := make([]string, 1)
	expectedHosts[0] = "ahostname.com" // nolint: goconst
	assert.Equal(t, manifest.Group{
		Name: "westcoast",
		Services: []manifest.Service{
			{
				Name:  "web",
				Image: "nginx",
				Resources: atypes.Resources{
					ID: 1,
					CPU: &atypes.CPU{
						Units: atypes.NewResourceValue(100),
					},
					GPU: &atypes.GPU{
						Units: atypes.NewResourceValue(1),
						Attributes: atypes.Attributes{
							{
								Key:   "vendor/nvidia/model/a100",
								Value: "true",
							},
						},
					},
					Memory: &atypes.Memory{
						Quantity: atypes.NewResourceValue(128 * unit.Mi),
					},
					Storage: atypes.Volumes{
						{
							Name:     "default",
							Quantity: atypes.NewResourceValue(1 * unit.Gi),
						},
					},
					Endpoints: []atypes.Endpoint{
						{
							Kind: atypes.Endpoint_SHARED_HTTP,
						},
						{
							Kind: atypes.Endpoint_RANDOM_PORT,
						},
					},
				},
				Count: 2,
				Expose: []manifest.ServiceExpose{
					{Port: 80, Global: true, Proto: manifest.TCP, Hosts: expectedHosts,
						HTTPOptions: manifest.ServiceExposeHTTPOptions{
							MaxBodySize: 1048576,
							ReadTimeout: 60000,
							SendTimeout: 60000,
							NextTries:   3,
							NextTimeout: 0,
							NextCases:   []string{"error", "timeout"},
						}},
					{Port: 12345, Global: true, Proto: manifest.UDP,
						HTTPOptions: manifest.ServiceExposeHTTPOptions{
							MaxBodySize: 1048576,
							ReadTimeout: 60000,
							SendTimeout: 60000,
							NextTries:   3,
							NextTimeout: 0,
							NextCases:   []string{"error", "timeout"},
						}},
				},
			},
		},
	}, mani.GetGroups()[0])
}

func TestV2_1_Parse_Deployments(t *testing.T) {
	sdl1, err := ReadFile("../x/deployment/testdata/deployment-v2.1.yaml")
	require.NoError(t, err)
	_, err = sdl1.DeploymentGroups()
	require.NoError(t, err)

	_, err = sdl1.Manifest()
	require.NoError(t, err)

	sha1, err := sdl1.Version()
	require.NoError(t, err)
	assert.Len(t, sha1, 32)

	sdl2, err := ReadFile("../x/deployment/testdata/deployment-v2.yaml")
	require.NoError(t, err)
	sha2, err := sdl2.Version()

	require.NoError(t, err)
	assert.Len(t, sha2, 32)
	require.NotEqual(t, sha1, sha2)
}

func Test_V2_1_Cross_Validates(t *testing.T) {
	sdl2, err := ReadFile("../x/deployment/testdata/deployment-v2.yaml")
	require.NoError(t, err)
	dgroups, err := sdl2.DeploymentGroups()
	require.NoError(t, err)
	m, err := sdl2.Manifest()
	require.NoError(t, err)

	// This is a single document producing both the manifest & deployment groups
	// These should always agree with each other. If this test fails at least one of the
	// following is ture
	// 1. Cross validation logic is wrong
	// 2. The DeploymentGroups() & Manifest() code do not agree with one another
	err = m.CheckAgainstGSpecs(dgroups)
	require.NoError(t, err)

	// Repeat the same test with another file
	sdl2, err = ReadFile("./_testdata/v2.1-simple.yaml")
	require.NoError(t, err)
	dgroups, err = sdl2.DeploymentGroups()
	require.NoError(t, err)
	m, err = sdl2.Manifest()
	require.NoError(t, err)

	// This is a single document producing both the manifest & deployment groups
	// These should always agree with each other
	err = m.CheckAgainstGSpecs(dgroups)
	require.NoError(t, err)

	// Repeat the same test with another file
	sdl2, err = ReadFile("./_testdata/v2.1-simple3.yaml")
	require.NoError(t, err)
	dgroups, err = sdl2.DeploymentGroups()
	require.NoError(t, err)
	m, err = sdl2.Manifest()
	require.NoError(t, err)

	// This is a single document producing both the manifest & deployment groups
	// These should always agree with each other
	err = m.CheckAgainstGSpecs(dgroups)
	require.NoError(t, err)

	// Repeat the same test with another file
	sdl2, err = ReadFile("./_testdata/v2.1-private_service.yaml")
	require.NoError(t, err)
	dgroups, err = sdl2.DeploymentGroups()
	require.NoError(t, err)
	m, err = sdl2.Manifest()
	require.NoError(t, err)

	// This is a single document producing both the manifest & deployment groups
	// These should always agree with each other
	err = m.CheckAgainstGSpecs(dgroups)
	require.NoError(t, err)

}

func Test_V2_1_Parse_simple(t *testing.T) {
	sdl, err := ReadFile("./_testdata/v2.1-simple.yaml")
	require.NoError(t, err)

	groups, err := sdl.DeploymentGroups()
	require.NoError(t, err)
	assert.Len(t, groups, 1)

	group := groups[0]
	assert.Len(t, group.GetResourceUnits(), 1)

	assert.Equal(t, atypes.Attribute{
		Key:   "region",
		Value: "us-west",
	}, group.Requirements.Attributes[0])

	assert.Len(t, group.GetResourceUnits(), 1)

	assert.Equal(t, dtypes.ResourceUnit{
		Count: 2,
		Resources: atypes.Resources{
			ID: 1,
			CPU: &atypes.CPU{
				Units: atypes.NewResourceValue(randCPU),
			},
			GPU: &atypes.GPU{
				Units: atypes.NewResourceValue(0),
			},
			Memory: &atypes.Memory{
				Quantity: atypes.NewResourceValue(randMemory),
			},
			Storage: atypes.Volumes{
				{
					Name:     "default",
					Quantity: atypes.NewResourceValue(randStorage),
				},
			},
			Endpoints: []atypes.Endpoint{
				{
					Kind: atypes.Endpoint_SHARED_HTTP,
				},
				{
					Kind: atypes.Endpoint_RANDOM_PORT,
				},
			},
		},
		Price: AkashDecCoin(t, 50),
	}, group.GetResourceUnits()[0])

	mani, err := sdl.Manifest()
	require.NoError(t, err)

	assert.Len(t, mani.GetGroups(), 1)

	expectedHosts := make([]string, 1)
	expectedHosts[0] = "ahostname.com"
	assert.Equal(t, manifest.Group{
		Name: "westcoast",
		Services: []manifest.Service{
			{
				Name:  "web",
				Image: "nginx",
				Resources: atypes.Resources{
					ID: 1,
					CPU: &atypes.CPU{
						Units: atypes.NewResourceValue(100),
					},
					GPU: &atypes.GPU{
						Units: atypes.NewResourceValue(0),
					},
					Memory: &atypes.Memory{
						Quantity: atypes.NewResourceValue(128 * unit.Mi),
					},
					Storage: atypes.Volumes{
						{
							Name:     "default",
							Quantity: atypes.NewResourceValue(1 * unit.Gi),
						},
					},
					Endpoints: []atypes.Endpoint{
						{
							Kind: atypes.Endpoint_SHARED_HTTP,
						},
						{
							Kind: atypes.Endpoint_RANDOM_PORT,
						},
					},
				},
				Count: 2,
				Expose: []manifest.ServiceExpose{
					{Port: 80, Global: true, Proto: manifest.TCP, Hosts: expectedHosts,
						HTTPOptions: manifest.ServiceExposeHTTPOptions{
							MaxBodySize: 1048576,
							ReadTimeout: 60000,
							SendTimeout: 60000,
							NextTries:   3,
							NextTimeout: 0,
							NextCases:   []string{"error", "timeout"},
						}},
					{Port: 12345, Global: true, Proto: manifest.UDP,
						HTTPOptions: manifest.ServiceExposeHTTPOptions{
							MaxBodySize: 1048576,
							ReadTimeout: 60000,
							SendTimeout: 60000,
							NextTries:   3,
							NextTimeout: 0,
							NextCases:   []string{"error", "timeout"},
						}},
				},
			},
		},
	}, mani.GetGroups()[0])

	assert.Nil(t, mani.GetGroups()[0].Services[0].Credentials)

}

func Test_V2_1_Parse_credentials(t *testing.T) {
	sdl, err := ReadFile("./_testdata/v2.1-credentials.yaml")
	require.NoError(t, err)

	mani, err := sdl.Manifest()
	require.NoError(t, err)

	assert.Len(t, mani.GetGroups(), 1)

	grp := mani.GetGroups()[0]
	assert.Len(t, grp.Services, 1)

	svc := grp.Services[0]

	assert.NotNil(t, svc)

	creds := svc.Credentials
	assert.NotNil(t, creds)

	assert.Equal(t, "https://test.com/v1", creds.Host)
	assert.Equal(t, "foo", creds.Username)
	assert.Equal(t, "foo", creds.Password)
}

func Test_V2_1_Parse_credentials_error(t *testing.T) {
	_, err := ReadFile("./_testdata/v2.1-credentials-error.yaml")
	require.Error(t, err)
}

func Test_v2_1_Parse_ProfileNameNotServiceName(t *testing.T) {
	sdl, err := ReadFile("./_testdata/v2.1-profile-svc-name-mismatch.yaml")
	require.NoError(t, err)

	dgroups, err := sdl.DeploymentGroups()
	require.NoError(t, err)
	assert.Len(t, dgroups, 1)

	mani, err := sdl.Manifest()
	require.NoError(t, err)
	assert.Len(t, mani.GetGroups(), 1)
}

func Test_v2_1_Parse_DeploymentNameServiceNameMismatch(t *testing.T) {
	sdl, err := ReadFile("./_testdata/v2.1-deployment-svc-mismatch.yaml")
	require.Error(t, err)
	require.Nil(t, sdl)
	require.Contains(t, err.Error(), "no service profile named")

	sdl, err = ReadFile("./_testdata/v2.1-simple2.yaml")
	require.NoError(t, err)
	require.NotNil(t, sdl)

	dgroups, err := sdl.DeploymentGroups()
	require.NoError(t, err)
	assert.Len(t, dgroups, 1)

	mani, err := sdl.Manifest()
	require.NoError(t, err)
	assert.Len(t, mani.GetGroups(), 1)

	require.Equal(t, dgroups[0].Name, mani.GetGroups()[0].Name)
	// SDL lists 2 services, but particular deployment specifies only one
	require.Len(t, mani.GetGroups()[0].Services, 1)

	// make sure deployment maps to the right service
	require.Len(t, mani.GetGroups()[0].Services[0].Expose, 2)
	require.Len(t, mani.GetGroups()[0].Services[0].Expose[0].Hosts, 1)
	require.Equal(t, mani.GetGroups()[0].Services[0].Expose[0].Hosts[0], "ahostname.com")
}

func TestV2_1_ParseServiceMix(t *testing.T) {
	sdl, err := ReadFile("./_testdata/v2.1-service-mix.yaml")
	require.NoError(t, err)

	groups, err := sdl.DeploymentGroups()
	require.NoError(t, err)
	assert.Len(t, groups, 1)

	group := groups[0]
	assert.Len(t, group.GetResourceUnits(), 2)
	assert.Len(t, group.Requirements.Attributes, 2)

	assert.Equal(t, atypes.Attribute{
		Key:   "region",
		Value: "us-west",
	}, group.Requirements.Attributes[1])

	assert.Equal(t, dtypes.ResourceUnits{
		{
			Count: 1,
			Resources: atypes.Resources{
				ID: 1,
				CPU: &atypes.CPU{
					Units: atypes.NewResourceValue(randCPU),
				},
				GPU: &atypes.GPU{
					Units: atypes.NewResourceValue(randGPU),
					Attributes: atypes.Attributes{
						{
							Key:   "vendor/nvidia/model/*",
							Value: "true",
						},
					},
				},
				Memory: &atypes.Memory{
					Quantity: atypes.NewResourceValue(randMemory),
				},
				Storage: atypes.Volumes{
					{
						Name:     "default",
						Quantity: atypes.NewResourceValue(randStorage),
					},
				},
				Endpoints: []atypes.Endpoint{
					{
						Kind: atypes.Endpoint_SHARED_HTTP,
					},
					{
						Kind: atypes.Endpoint_RANDOM_PORT,
					},
				},
			},
			Price: AkashDecCoin(t, 50),
		},
		{
			Count: 1,
			Resources: atypes.Resources{
				ID: 2,
				CPU: &atypes.CPU{
					Units: atypes.NewResourceValue(randCPU),
				},
				GPU: &atypes.GPU{
					Units: atypes.NewResourceValue(0),
				},
				Memory: &atypes.Memory{
					Quantity: atypes.NewResourceValue(randMemory),
				},
				Storage: atypes.Volumes{
					{
						Name:     "default",
						Quantity: atypes.NewResourceValue(randStorage),
					},
				},
				Endpoints: []atypes.Endpoint{
					{
						Kind: atypes.Endpoint_SHARED_HTTP,
					},
					{
						Kind: atypes.Endpoint_RANDOM_PORT,
					},
				},
			},
			Price: AkashDecCoin(t, 50),
		},
	}, group.GetResourceUnits())

	mani, err := sdl.Manifest()
	require.NoError(t, err)

	assert.Len(t, mani.GetGroups(), 1)

	assert.Equal(t, manifest.Group{
		Name: "westcoast",
		Services: []manifest.Service{
			{
				Name:  "svca",
				Image: "nginx",
				Resources: atypes.Resources{
					ID: 1,
					CPU: &atypes.CPU{
						Units: atypes.NewResourceValue(100),
					},
					GPU: &atypes.GPU{
						Units: atypes.NewResourceValue(1),
						Attributes: atypes.Attributes{
							{
								Key:   "vendor/nvidia/model/*",
								Value: "true",
							},
						},
					},
					Memory: &atypes.Memory{
						Quantity: atypes.NewResourceValue(128 * unit.Mi),
					},
					Storage: atypes.Volumes{
						{
							Name:     "default",
							Quantity: atypes.NewResourceValue(1 * unit.Gi),
						},
					},
					Endpoints: []atypes.Endpoint{
						{
							Kind: atypes.Endpoint_SHARED_HTTP,
						},
						{
							Kind: atypes.Endpoint_RANDOM_PORT,
						},
					},
				},
				Count: 1,
				Expose: []manifest.ServiceExpose{
					{
						Port: 80, Global: true, Proto: manifest.TCP, Hosts: []string{"ahostname.com"},
						HTTPOptions: defaultHTTPOptions,
					},
					{
						Port: 12345, Global: true, Proto: manifest.UDP,
						HTTPOptions: defaultHTTPOptions,
					},
				},
			},
			{
				Name:  "svcb",
				Image: "nginx",
				Resources: atypes.Resources{
					ID: 2,
					CPU: &atypes.CPU{
						Units: atypes.NewResourceValue(100),
					},
					GPU: &atypes.GPU{
						Units: atypes.NewResourceValue(0),
					},
					Memory: &atypes.Memory{
						Quantity: atypes.NewResourceValue(128 * unit.Mi),
					},
					Storage: atypes.Volumes{
						{
							Name:     "default",
							Quantity: atypes.NewResourceValue(1 * unit.Gi),
						},
					},
					Endpoints: []atypes.Endpoint{
						{
							Kind: atypes.Endpoint_SHARED_HTTP,
						},
						{
							Kind: atypes.Endpoint_RANDOM_PORT,
						},
					},
				},
				Count: 1,
				Expose: []manifest.ServiceExpose{
					{
						Port: 80, Global: true, Proto: manifest.TCP, Hosts: []string{"bhostname.com"},
						HTTPOptions: defaultHTTPOptions,
					},
					{
						Port: 12346, Global: true, Proto: manifest.UDP,
						HTTPOptions: defaultHTTPOptions,
					},
				},
			},
		},
	}, mani.GetGroups()[0])
}

func TestV2_1_ParseServiceMix2(t *testing.T) {
	sdl, err := ReadFile("./_testdata/v2.1-service-mix2.yaml")
	require.NoError(t, err)

	groups, err := sdl.DeploymentGroups()
	require.NoError(t, err)
	assert.Len(t, groups, 1)

	group := groups[0]
	assert.Len(t, group.GetResourceUnits(), 1)
	assert.Len(t, group.Requirements.Attributes, 2)

	assert.Equal(t, atypes.Attribute{
		Key:   "region",
		Value: "us-west",
	}, group.Requirements.Attributes[1])

	assert.Equal(t, dtypes.ResourceUnits{
		{
			Count: 2,
			Resources: atypes.Resources{
				ID: 1,
				CPU: &atypes.CPU{
					Units: atypes.NewResourceValue(randCPU),
				},
				GPU: &atypes.GPU{
					Units: atypes.NewResourceValue(randGPU),
					Attributes: atypes.Attributes{
						{
							Key:   "vendor/nvidia/model/*",
							Value: "true",
						},
					},
				},
				Memory: &atypes.Memory{
					Quantity: atypes.NewResourceValue(randMemory),
				},
				Storage: atypes.Volumes{
					{
						Name:     "default",
						Quantity: atypes.NewResourceValue(randStorage),
					},
				},
				Endpoints: []atypes.Endpoint{
					{
						Kind: atypes.Endpoint_SHARED_HTTP,
					},
					{
						Kind: atypes.Endpoint_RANDOM_PORT,
					},
					{
						Kind: atypes.Endpoint_SHARED_HTTP,
					},
					{
						Kind: atypes.Endpoint_RANDOM_PORT,
					},
				},
			},
			Price: AkashDecCoin(t, 50),
		},
	}, group.GetResourceUnits())

	mani, err := sdl.Manifest()
	require.NoError(t, err)

	assert.Len(t, mani.GetGroups(), 1)

	assert.Equal(t, manifest.Group{
		Name: "westcoast",
		Services: []manifest.Service{
			{
				Name:  "svca",
				Image: "nginx",
				Resources: atypes.Resources{
					ID: 1,
					CPU: &atypes.CPU{
						Units: atypes.NewResourceValue(100),
					},
					GPU: &atypes.GPU{
						Units: atypes.NewResourceValue(1),
						Attributes: atypes.Attributes{
							{
								Key:   "vendor/nvidia/model/*",
								Value: "true",
							},
						},
					},
					Memory: &atypes.Memory{
						Quantity: atypes.NewResourceValue(128 * unit.Mi),
					},
					Storage: atypes.Volumes{
						{
							Name:     "default",
							Quantity: atypes.NewResourceValue(1 * unit.Gi),
						},
					},
					Endpoints: []atypes.Endpoint{
						{
							Kind: atypes.Endpoint_SHARED_HTTP,
						},
						{
							Kind: atypes.Endpoint_RANDOM_PORT,
						},
					},
				},
				Count: 1,
				Expose: []manifest.ServiceExpose{
					{
						Port: 80, Global: true, Proto: manifest.TCP, Hosts: []string{"ahostname.com"},
						HTTPOptions: defaultHTTPOptions,
					},
					{
						Port: 12345, Global: true, Proto: manifest.UDP,
						HTTPOptions: defaultHTTPOptions,
					},
				},
			},
			{
				Name:  "svcb",
				Image: "nginx",
				Resources: atypes.Resources{
					ID: 1,
					CPU: &atypes.CPU{
						Units: atypes.NewResourceValue(100),
					},
					GPU: &atypes.GPU{
						Units: atypes.NewResourceValue(1),
						Attributes: atypes.Attributes{
							{
								Key:   "vendor/nvidia/model/*",
								Value: "true",
							},
						},
					},
					Memory: &atypes.Memory{
						Quantity: atypes.NewResourceValue(128 * unit.Mi),
					},
					Storage: atypes.Volumes{
						{
							Name:     "default",
							Quantity: atypes.NewResourceValue(1 * unit.Gi),
						},
					},
					Endpoints: []atypes.Endpoint{
						{
							Kind: atypes.Endpoint_SHARED_HTTP,
						},
						{
							Kind: atypes.Endpoint_RANDOM_PORT,
						},
					},
				},
				Count: 1,
				Expose: []manifest.ServiceExpose{
					{
						Port: 80, Global: true, Proto: manifest.TCP, Hosts: []string{"bhostname.com"},
						HTTPOptions: defaultHTTPOptions,
					},
					{
						Port: 12346, Global: true, Proto: manifest.UDP,
						HTTPOptions: defaultHTTPOptions,
					},
				},
			},
		},
	}, mani.GetGroups()[0])
}
