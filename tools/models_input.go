package tools

import (
	"strings"

	"github.com/dpecos/cbox/models"
	"github.com/dpecos/cbox/tools/console"
)

func ConsoleReadCommand() *models.Command {

	command := models.Command{
		ID:          console.ReadString("ID"),
		Title:       console.ReadString("Title"),
		Description: console.ReadStringMulti("Description"),
		URL:         console.ReadString("URL"),
		Code:        console.ReadStringMulti("Code / Command"),
		Tags:        []string{},
	}
	tags := console.ReadString("Tags (separated by space)")
	for _, tag := range strings.Split(tags, " ") {
		if tag != "" {
			command.Tags = append(command.Tags, tag)
		}
	}

	return &command
}

func ConsoleEditCommand(command *models.Command) {
	command.ID = console.EditString("ID", command.ID)
	command.Title = console.EditString("Title", command.Title)
	command.Description = console.EditStringMulti("Description", command.Description)
	command.URL = console.EditString("URL", command.URL)
	command.Code = console.EditStringMulti("Code / Command", command.Code)
}