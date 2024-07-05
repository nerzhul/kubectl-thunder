package main

import (
	"log"
	"os"

	"github.com/nerzhul/kubectl-thunder/pkg/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "thunder",
		Version: "0.1.0",
		Usage:   "An enhanced CLI for Kubernetes",
		Commands: []*cli.Command{
			{
				Name: "nodes",
				Subcommands: []*cli.Command{
					{
						Name:        "check-allocation",
						Description: "Check if nodes can allocate the specified resources",
						Action: func(c *cli.Context) error {
							return commands.Nodes_can_allocate(c)
						},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "memory",
								Usage:    "Memory resource to try to allocate",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "cpu",
								Usage:    "CPU resource to try to allocate",
								Required: true,
							},
							&cli.StringSliceFlag{
								Name:  "show-labels",
								Usage: "Print node labels",
							},
						},
					},
				},
			},
			{
				Name: "secrets",
				Subcommands: []*cli.Command{{
					Name: "find",
					Subcommands: []*cli.Command{
						{
							Name:        "expiring",
							Description: "Find expiring certificates in the cluster secrets",
							Action: func(c *cli.Context) error {
								return commands.Secrets_find_expiring_certificates(c)
							},
							Flags: []cli.Flag{
								&cli.BoolFlag{
									Name:  "used-only",
									Usage: "Only report certificates that are used by ingresses",
								},
								&cli.BoolFlag{
									Name:  "unused-only",
									Usage: "Only report certificates that are not used by ingresses",
								},
								&cli.BoolFlag{
									Name:  "delete",
									Usage: "Delete expired certificates after reporting them (use with caution!)",
								},
								&cli.Uint64Flag{
									Name:        "after",
									Usage:       "Only consider certificates that are expired after the specified number of days",
									DefaultText: "0",
								},
							},
							Before: func(c *cli.Context) error {
								if c.IsSet("used-only") && c.IsSet("unused-only") {
									return cli.Exit("used-only and unused-only flags are mutually exclusive", 1)
								}

								if c.IsSet("delete") && c.IsSet("after") {
									return cli.Exit("delete and after flags are mutually exclusive", 1)
								}

								return nil
							},
						},
						{
							Name:        "by-certificate-san",
							Description: "Find certificates secrets by SAN",
							Action: func(c *cli.Context) error {
								return commands.Secrets_find_certificates_by_san(c)
							},
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:     "san",
									Usage:    "Subject Alternative Name to search for",
									Required: true,
								},
								&cli.BoolFlag{
									Name:  "wildcard-match",
									Usage: "Subject Alternative Name is matched by a wildcard certificate",
								},
							},
						},
					},
				}},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
