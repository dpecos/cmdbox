package controllers

import (
	"fmt"
	"log"
	"strings"

	"github.com/dplabs/cbox/src/core"
	"github.com/dplabs/cbox/src/models"
	"github.com/dplabs/cbox/src/tools"
	"github.com/dplabs/cbox/src/tools/console"
)

func (ctrl *CLIController) CloudSpaceInfo(args []string) {
	console.PrintAction("Retrieving info of an space")

	selector, err := models.ParseSelectorForCloud(args[0])
	if err != nil {
		log.Fatalf("ctrl.cloud: space info: %v", err)
	}

	space, err := ctrl.cloud.SpaceFind(selector)
	if err != nil {
		log.Fatalf("ctrl.cloud: space info: %v", err)
	}

	tools.PrintSpace(selector.String(), space)
}

func (ctrl *CLIController) CloudSpacePublish(args []string) {
	console.PrintAction("Publishing an space")

	selector, err := models.ParseSelectorMandatorySpace(args[0])
	if err != nil {
		log.Fatalf("ctrl.cloud: publish space: %v", err)
	}

	space, err := ctrl.findSpace(selector)
	if err != nil {
		log.Fatalf("ctrl.cloud: publish space: %v", err)
	}

	if space.Selector.Namespace == "" {
		space.Selector.NamespaceType = models.TypeUser
		space.Selector.Namespace = ctrl.cloud.Login
	}

	previousSpace := space.Selector.Namespace

	if OrganizationOption != "" {
		space.Selector.NamespaceType = models.TypeOrganization
		space.Selector.Namespace = OrganizationOption
	}

	tools.PrintSpace("Space to publish", space)

	if selector.Item != "" {
		commands := space.CommandList(selector.Item)
		if len(commands) == 0 {
			log.Fatalf("ctrl.cloud: no local commands matched selector: %s", selector)
		}

		space.Entries = commands
	}

	// tools.PrintCommandList("Containing these commands", space.Entries, false, false)

	if OrganizationOption != "" && previousSpace != OrganizationOption {
		console.PrintWarning(fmt.Sprintf("You're about to publish workspace '%s' under a different organization '%s'\n", space.String(), OrganizationOption))
	}

	if SkipQuestionsFlag || console.Confirm("Publish?") {
		fmt.Printf("Publishing space '%s'...\n", space.String())

		err = ctrl.cloud.SpacePublish(space)
		if err != nil {
			log.Fatalf("ctrl.cloud: publish space: %v", err)
		}

		ctrl.cleanOldSpaceFile(space, selector)

		core.Save(ctrl.cbox) // to store space's new namespace

		console.PrintSuccess("Space published successfully!")
	} else {
		console.PrintError("Publishing cancelled")
	}
}

func (ctrl *CLIController) CloudSpaceUnpublish(args []string) {
	console.PrintAction("Unpublishing an space")

	selector, err := models.ParseSelectorForCloud(args[0])
	if err != nil {
		log.Fatalf("ctrl.cloud: unpublish space: %v", err)
	}

	tools.PrintSelector("Space to unpublish", selector)

	_, err = ctrl.findSpace(selector)
	if err == nil {
		console.PrintInfo("Local copy won't be deleted")
	} else {
		console.PrintWarning("You don't have a local copy of the space")
	}

	if SkipQuestionsFlag || console.Confirm("Unpublish?") {
		fmt.Printf("Unpublishing space '%s'...\n", selector.String())

		err = ctrl.cloud.SpaceUnpublish(selector)
		if err != nil {
			log.Fatalf("ctrl.cloud: unpublish space: %v", err)
		}

		console.PrintSuccess("Space unpublished successfully!")
	} else {
		console.PrintError("Unpublishing cancelled")
	}
}

func (ctrl *CLIController) CloudSpaceClone(args []string) {
	console.PrintAction("Cloning an space")

	selector, err := models.ParseSelectorForCloud(args[0])
	if err != nil {
		log.Fatalf("ctrl.cloud: clone space: invalid ctrl.cloud selector: %v", err)
	}

	space, err := ctrl.cloud.SpaceFind(selector)
	if err != nil {
		log.Fatalf("ctrl.cloud: clone space: %v", err)
	}

	commands, err := ctrl.cloud.CommandList(selector)
	if err != nil {
		log.Fatalf("ctrl.cloud: list commands: %v", err)
	}

	space.Entries = commands

	tools.PrintSpace("Space to clone", space)
	tools.PrintCommandList("Containing these commands", space.Entries, false, false)

	if SkipQuestionsFlag || console.Confirm("Clone?") {
		err := ctrl.cbox.SpaceCreate(space)
		for err != nil {
			console.PrintError("Space already found in your cbox. Try a different one")
			space.Label = strings.ToLower(console.ReadString("Label", console.NOT_EMPTY_VALUES, console.ONLY_VALID_CHARS))
			err = ctrl.cbox.SpaceCreate(space)
		}

		core.Save(ctrl.cbox)

		console.PrintSuccess("Space cloned successfully!")
	} else {
		console.PrintError("Clone cancelled")
	}
}

// TODO: needed?
func (ctrl *CLIController) CloudSpacePull(args []string) {
	console.PrintAction("Pulling latest changes of an space")

	selector, err := models.ParseSelectorMandatorySpace(args[0])
	if err != nil {
		log.Fatalf("ctrl.cloud: pull space: invalid ctrl.cloud selector: %v", err)
	}

	space, err := ctrl.findSpace(selector)
	if err != nil {
		log.Fatalf("ctrl.cloud: pull space: %v", err)
	}

	spaceCloud, err := ctrl.cloud.SpaceFind(selector)
	if err != nil {
		log.Fatalf("ctrl.cloud: pull space: %v", err)
	}

	commands, err := ctrl.cloud.CommandList(selector)
	if err != nil {
		log.Fatalf("ctrl.cloud: list commands: %v", err)
	}

	// Note: Label is not overwritten because user can renamed his local copy of the space
	space.Entries = commands
	space.UpdatedAt = spaceCloud.UpdatedAt
	space.Description = spaceCloud.Description

	core.Save(ctrl.cbox)

	tools.PrintSpace("Pulled space", space)

	console.PrintSuccess("Space pulled successfully!")
}