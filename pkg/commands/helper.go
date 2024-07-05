package commands

import "github.com/urfave/cli/v2"

func boolArg(c *cli.Context, name string) bool {
	return c.IsSet(name) && c.Bool(name)
}

func u64Arg(c *cli.Context, name string) uint64 {
	if c.IsSet(name) {
		return c.Uint64(name)
	}

	return 0
}