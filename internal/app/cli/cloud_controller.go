package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/dplabs/cbox/internal/app/core"
	"github.com/dplabs/cbox/internal/pkg"
	"github.com/dplabs/cbox/internal/pkg/console"
	"github.com/dplabs/cbox/pkg/models"
	"github.com/spf13/cobra"
)

func (ctrl *CLIController) CloudLogin(cmd *cobra.Command, args []string) {
	fmt.Println(pkg.Logo)
	url := fmt.Sprintf("%s/auth/", core.CloudURL())
	fmt.Printf("Open this URL in a browser and follow the authentication process: \n\n%s\n\n", url)

	jwt := console.ReadString("JWT Token", console.NOT_EMPTY_VALUES)
	fmt.Println()

	_, _, name, err := core.CloudLogin(jwt)
	if err != nil {
		console.PrintError("Error trying to parse JWT token. Try to login again")
		log.Fatalf("cloud: login: %v", err)
	}

	console.PrintSuccess("Hi " + name + "!")
}

func (ctrl *CLIController) CloudLogout(cmd *cobra.Command, args []string) {
	fmt.Println(pkg.Logo)
	core.CloudLogout()
	console.PrintSuccess("Successfully logged out from cbox cloud. See you back soon!")
}

func (ctrl *CLIController) CloudSpacePublish(cmd *cobra.Command, args []string) {
	console.PrintAction("Publishing an space")

	selector, err := models.ParseSelectorMandatorySpace(args[0])
	if err != nil {
		log.Fatalf("cloud: publish item: %v", err)
	}

	space, err := cboxInstance.SpaceFind(selector.Space)
	if err != nil {
		log.Fatalf("cloud: publish item: %v", err)
	}

	pkg.PrintSpace("Space to publish", space)

	if selector.Item != "" {
		commands := space.CommandList(selector.Item)
		if len(commands) == 0 {
			log.Fatalf("cloud: no local commands matched selector: %s", selector)
		}

		space.Entries = commands
	}

	// pkg.PrintCommandList("Containing these commands", space.Entries, false, false)

	if console.Confirm("Publish?") {
		fmt.Printf("Publishing space '%s'...\n", space.Label)

		cloud, err := core.CloudClient()
		if err != nil {
			log.Fatalf("cloud: publish space: %v", err)
		}
		err = cloud.SpacePublish(space)
		if err != nil {
			log.Fatalf("cloud: publish space: %v", err)
		}

		console.PrintSuccess("Space published successfully!")
	} else {
		console.PrintError("Publishing cancelled")
	}
}

func (ctrl *CLIController) CloudSpaceUnpublish(cmd *cobra.Command, args []string) {
	console.PrintAction("Unpublishing an space")

	selector, err := models.ParseSelectorMandatorySpace(args[0])
	if err != nil {
		log.Fatalf("cloud: unpublish item: %v", err)
	}

	pkg.PrintSelector("Space to unpublish", selector)

	_, err = cboxInstance.SpaceFind(selector.Space)
	if err == nil {
		console.PrintInfo("Local copy won't be deleted")
	} else {
		console.PrintWarning("You don't have a local copy of the space")
	}

	if console.Confirm("Unpublish?") {
		fmt.Printf("Unpublishing space '%s'...\n", selector.String())

		cloud, err := core.CloudClient()
		if err != nil {
			log.Fatalf("cloud: unpublish space: %v", err)
		}
		err = cloud.SpaceUnpublish(selector)
		if err != nil {
			log.Fatalf("cloud: unpublish space: %v", err)
		}

		console.PrintSuccess("Space unpublished successfully!")
	} else {
		console.PrintError("Unpublishing cancelled")
	}
}

func (ctrl *CLIController) CloudSpaceClone(cmd *cobra.Command, args []string) {
	console.PrintAction("Cloning an space")

	selector, err := models.ParseSelectorForCloudCommand(args[0])
	if err != nil {
		log.Fatalf("cloud: clone space: invalid cloud selector: %v", err)
	}

	cloud, err := core.CloudClient()
	if err != nil {
		log.Fatalf("cloud: clone space: cloud client: %v", err)
	}

	space, err := cloud.SpaceRetrieve(selector, nil)
	if err != nil {
		log.Fatalf("cloud: clone space: %v", err)
	}

	pkg.PrintSpace("Space to clone", space)
	pkg.PrintCommandList("Containing these commands", space.Entries, false, false)

	if console.Confirm("Clone?") {
		err := cboxInstance.SpaceCreate(space)
		for err != nil {
			console.PrintError("Space already found in your cbox. Try a different one")
			space.Label = strings.ToLower(console.ReadString("Label", console.NOT_EMPTY_VALUES, console.ONLY_VALID_CHARS))
			err = cboxInstance.SpaceCreate(space)
		}

		core.Save(cboxInstance)

		console.PrintSuccess("Space cloned successfully!")
	} else {
		console.PrintError("Clone cancelled")
	}
}

func (ctrl *CLIController) CloudSpacePull(cmd *cobra.Command, args []string) {
	console.PrintAction("Pulling latest changes of an space")

	selector, err := models.ParseSelectorMandatorySpace(args[0])
	if err != nil {
		log.Fatalf("cloud: pull space: invalid cloud selector: %v", err)
	}

	cloud, err := core.CloudClient()
	if err != nil {
		log.Fatalf("cloud: pull space: cloud client: %v", err)
	}

	space, err := cboxInstance.SpaceFind(selector.Space)
	if err != nil {
		log.Fatalf("cloud: pull space: %v", err)
	}

	spaceCloud, err := cloud.SpaceRetrieve(nil, &space.ID)
	if err != nil {
		log.Fatalf("cloud: pull space: %v", err)
	}

	// Note: Label is not overwritten because user can renamed his local copy of the space
	space.Entries = spaceCloud.Entries
	space.UpdatedAt = spaceCloud.UpdatedAt
	space.Description = spaceCloud.Description

	core.Save(cboxInstance)

	pkg.PrintSpace("Pulled space", space)

	console.PrintSuccess("Space pulled successfully!")
}

func (ctrl *CLIController) CloudCommandList(cmd *cobra.Command, args []string) {
	selector, err := models.ParseSelectorForCloudCommand(args[0])
	if err != nil {
		log.Fatalf("cloud: list commands: invalid cloud selector: %v", err)
	}

	cloud, err := core.CloudClient()
	if err != nil {
		log.Fatalf("cloud: list commands: cloud client: %v", err)
	}

	commands, err := cloud.CommandList(selector)
	if err != nil {
		log.Fatalf("cloud: list commands: %v", err)
	}

	pkg.PrintCommandList(selector.String(), commands, viewSnippet, false)
}

func (ctrl *CLIController) CloudCommandCopy(cmd *cobra.Command, args []string) {
	console.PrintAction("Copying cloud commands")

	cmdSelector, err := models.ParseSelectorForCloudCommand(args[0])
	if err != nil {
		log.Fatalf("cloud: copy command: invalid cloud selector: %v", err)
	}

	spaceSelector, err := models.ParseSelectorMandatorySpace(args[1])
	if err != nil {
		log.Fatalf("cloud: copy command: invalid space selector: %v", err)
	}

	space, err := cboxInstance.SpaceFind(spaceSelector.Space)
	if err != nil {
		log.Fatalf("cloud: copy command: local space: %v", err)
	}

	cloud, err := core.CloudClient()
	if err != nil {
		log.Fatalf("cloud: copy command: cloud client: %v", err)
	}

	commands, err := cloud.CommandList(cmdSelector)
	if err != nil {
		log.Fatalf("cloud: copy command: retrieving matches: %v", err)
	}

	if len(commands) == 0 {
		console.PrintError(fmt.Sprintf("Command '%s' not found", cmdSelector))
	}

	pkg.PrintCommandList("Commands to copy", commands, false, false)

	if console.Confirm(fmt.Sprintf("Copy these commands into %s?", spaceSelector)) {

		failures := false
		for _, command := range commands {
			err = space.CommandAdd(&command)
			if err != nil {
				failures = true
				log.Printf("cloud: copy command: %v", err)
			}
		}

		core.Save(cboxInstance)

		if failures {
			console.PrintError("Some commands could not be stored")
		} else {
			console.PrintSuccess("Commands copied successfully!")
		}
	} else {
		console.PrintError("Copy cancelled")
	}
}
