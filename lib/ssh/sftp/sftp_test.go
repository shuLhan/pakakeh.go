package sftp

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/ssh"
	"github.com/shuLhan/share/lib/ssh/config"
	"github.com/shuLhan/share/lib/test"
)

const (
	envNameTestManual = "LIB_SFTP_TEST_MANUAL"
)

var (
	// testClient the sftp Client to test all exported functionalities,
	// and also to test concurrent packet communication.
	testClient *Client

	// Flag to run the unit test that require SSH server.
	// This flag is set through environment variable defined on
	// envNameTestManual.
	isTestManual bool
)

func TestMain(m *testing.M) {
	isTestManual = len(os.Getenv(envNameTestManual)) > 0
	if !isTestManual {
		return
	}

	cfg := &config.Section{
		Field: map[string]string{
			config.KeyUser:     `ms`,
			config.KeyHostname: `127.0.0.1`,
			config.KeyPort:     `22`,
		},
		IdentityFile: []string{
			"./testdata/id_ed25519",
		},
	}

	sshClient, err := ssh.NewClientFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	testClient, err = NewClient(sshClient.Client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server version: %d\n", testClient.version)
	fmt.Printf("Server extensions: %v\n", testClient.exts)

	os.Exit(m.Run())
}

func TestErrFailure(t *testing.T) {
	err := errFailure("test")
	got := errors.Is(err, ErrFailure)
	test.Assert(t, "ErrFailure", true, got)
}
