package container

import (
	"os"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

func TestContainerAuth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		username  string
		password  string
		registry  string
		expectErr bool
		expectMsg string
	}{
		{
			name:      "valid credentials for docker login",
			username:  "admin",
			password:  "admin",
			registry:  "registry.faas.host",
			expectErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := os.Setenv("DOCKER_REGISTRY", tc.registry)
			require.NoError(t, err)
			err = os.Setenv("DOCKER_USERNAME", tc.username)
			require.NoError(t, err)
			err = os.Setenv("DOCKER_PASSWORD", tc.password)
			require.NoError(t, err)
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				panic(err)
			}
			defer cli.Close()
			err = Auth(t.Context(), cli)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectMsg)
			} else {
				require.NoError(t, err)
			}
			t.Cleanup(func() {
				err := os.Unsetenv("DOCKER_REGISTRY")
				require.NoError(t, err)
			})
		})
	}
}
