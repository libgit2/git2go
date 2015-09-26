package git

import (
	"os"
	"testing"
)

func setupConfig() (*Config, error) {
	var (
		c   *Config
		err error
		p   string
	)

	p, err = ConfigFindGlobal()
	if err != nil {
		return nil, err
	}

	c, err = OpenOndisk(nil, p)
	if err != nil {
		return nil, err
	}

	c.SetString("foo.bar", "baz")

	return c, err
}

func cleanupConfig() {
	os.Remove(tempConfig)
}

func TestConfigLookupString(t *testing.T) {
	var (
		err error
		val string
		c   *Config
	)

	c, err = setupConfig()
	defer cleanupConfig()
	if err != nil {
		t.Errorf("Setup error: '%v'. Expected none\n", err)
		t.FailNow()
	}
	defer c.Free()

	val, err = c.LookupString("foo.bar")
	if err != nil {
		t.Errorf("Got error: '%v', expected none\n", err)
		t.FailNow()
	}

	if val != "baz" {
		t.Errorf("Got '%s', expected 'bar'\n", val)
	}
}
