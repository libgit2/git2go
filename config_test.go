package git

import (
	"os"
	"testing"
)

var tempConfig = "./temp.gitconfig"

func setupConfig() (*Config, error) {
	var (
		c   *Config
		err error
	)

	c, err = OpenOndisk(nil, tempConfig)
	if err != nil {
		return nil, err
	}

	err = c.SetString("foo.bar", "baz")
	if err != nil {
		return nil, err
	}
	err = c.SetBool("foo.bool", true)
	if err != nil {
		return nil, err
	}
	err = c.SetInt32("foo.int32", 32)
	if err != nil {
		return nil, err
	}
	err = c.SetInt64("foo.int64", 64)
	if err != nil {
		return nil, err
	}

	return c, err
}

func cleanupConfig() {
	os.Remove(tempConfig)
}

type TestRunner func(*Config, *testing.T)

var tests = []TestRunner{
	// LookupString
	func(c *Config, t *testing.T) {
		val, err := c.LookupString("foo.bar")
		if err != nil {
			t.Errorf("Got LookupString error: '%v', expected none\n", err)
		}
		if val != "baz" {
			t.Errorf("Got '%s' from LookupString, expected 'bar'\n", val)
		}
	},
	// LookupBool
	func(c *Config, t *testing.T) {
		val, err := c.LookupBool("foo.bool")
		if err != nil {
			t.Errorf("Got LookupBool error: '%v', expected none\n", err)
		}
		if !val {
			t.Errorf("Got %t from LookupBool, expected 'false'\n", val)
		}
	},
	// LookupInt32
	func(c *Config, t *testing.T) {
		val, err := c.LookupInt32("foo.int32")
		if err != nil {
			t.Errorf("Got LookupInt32 error: '%v', expected none\n", err)
		}
		if val != 32 {
			t.Errorf("Got %v, expected 32\n", val)
		}
	},
	// LookupInt64
	func(c *Config, t *testing.T) {
		val, err := c.LookupInt64("foo.int64")
		if err != nil {
			t.Errorf("Got LookupInt64 error: '%v', expected none\n", err)
		}
		if val != 64 {
			t.Errorf("Got %v, expected 64\n", val)
		}
	},
}

func TestConfigLookups(t *testing.T) {
	t.Parallel()
	var (
		err error
		c   *Config
	)

	c, err = setupConfig()
	defer cleanupConfig()

	if err != nil {
		t.Errorf("Setup error: '%v'. Expected none\n", err)
		return
	}
	defer c.Free()

	for _, test := range tests {
		test(c, t)
	}
}
