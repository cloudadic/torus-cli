package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/promptui"
)

func init() {
	services := cli.Command{
		Name:     "services",
		Usage:    "View and manipulate services within an organization",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List services for an organization",
				Flags: []cli.Flag{
					OrgFlag("org to show services for", true),
					ProjectFlag("project to show services for", false),
					cli.BoolFlag{
						Name:  "all",
						Usage: "Perform command on all projects",
					},
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, listServices,
				),
			},
			{
				Name:      "create",
				Usage:     "Create a service in an organization",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					OrgFlag("Create the project in this org", false),
					ProjectFlag("project to create services for", false),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					createServiceCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, services)
}

const serviceListFailed = "Could not list services, please try again."

func listServices(ctx *cli.Context) error {
	if !ctx.Bool("all") {
		if len(ctx.String("project")) < 1 {
			text := "Missing flags: --project\n\n"
			text += usageString(ctx)
			return cli.NewExitError(text, -1)
		}
	} else {
		if len(ctx.String("project")) > 0 {
			text := "Cannot use --project flag with --all\n\n"
			text += usageString(ctx)
			return cli.NewExitError(text, -1)
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	var org *api.OrgResult
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return cli.NewExitError(serviceListFailed, -1)
	}
	if org == nil {
		return cli.NewExitError("Org not found.", -1)
	}

	// Identify which projects to list services for
	var projectID identity.ID
	var projects []api.ProjectResult
	if ctx.Bool("all") {
		// Pull all projects for the given orgID
		projects, err = client.Projects.List(c, org.ID, nil)
		if err != nil {
			return cli.NewExitError(serviceListFailed, -1)
		}

	} else {
		// Retrieve only a single project by name
		projectName := ctx.String("project")
		projects, err = client.Projects.List(c, org.ID, &projectName)
		if err != nil {
			return cli.NewExitError(serviceListFailed, -1)
		}
		if len(projects) == 1 {
			projectID = *projects[0].ID
		} else {
			return cli.NewExitError("Project not found.", -1)
		}
	}

	// Retrieve services for targeted org and project
	var services []api.ServiceResult
	services, err = client.Services.List(c, org.ID, &projectID, nil)
	if err != nil {
		return cli.NewExitError(serviceListFailed, -1)
	}

	// Build map of services to project
	pMap := make(map[string]api.ProjectResult)
	for _, project := range projects {
		pMap[project.ID.String()] = project
	}
	sMap := make(map[string][]api.ServiceResult)
	for _, service := range services {
		ID := service.Body.ProjectID.String()
		sMap[ID] = append(sMap[ID], service)
	}

	// Build output of projects/services
	fmt.Println("")
	for projectID, project := range pMap {
		count := strconv.Itoa(len(sMap[projectID]))
		title := project.Body.Name + " (" + count + ")"
		fmt.Println(title)
		fmt.Println(strings.Repeat("-", utf8.RuneCountInString(title)))
		for _, service := range sMap[projectID] {
			fmt.Println(service.Body.Name)
		}
		fmt.Println("")
	}

	return nil
}

const serviceCreateFailed = "Could not create service. Please try again."

func createServiceCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return cli.NewExitError(serviceCreateFailed, -1)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, project, createdNew, err := SelectCreateOrgAndProject(client, c, ctx, ctx.String("org"), ctx.String("project"))
	if err != nil {
		if err != promptui.ErrEOF && err != promptui.ErrInterrupt {
			fmt.Println("")
		}
		return err
	}
	if org == nil {
		fmt.Println("")
		return cli.NewExitError("Org not found", -1)
	}
	if project == nil {
		fmt.Println("")
		return cli.NewExitError("Project not found", -1)
	}
	if createdNew {
		fmt.Println("")
	}

	args := ctx.Args()
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	label := "Service name"
	if name == "" {
		name, err = NamePrompt(&label, "")
		if err != nil {
			if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
				return err
			}
			fmt.Println("")
			return cli.NewExitError(serviceCreateFailed, -1)
		}
	} else {
		fmt.Println(promptui.SuccessfulValue(label, name))
	}

	fmt.Println("")
	err = client.Services.Create(c, org.ID, project.ID, name)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return cli.NewExitError("Service already exists", -1)
		}
		fmt.Printf("%v\n", err)
		return cli.NewExitError(serviceCreateFailed, -1)
	}

	fmt.Printf("Service %s created.\n", name)
	return nil
}
