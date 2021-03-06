package models

import "strings"

func (command *Command) TagAdd(tag string) {
	if tag == "" {
		return
	}

	found := false

	tag = strings.ToLower(tag)

	for _, t := range command.Tags {
		if t == tag {
			found = true
			break
		}
	}

	if !found {
		command.Tags = append(command.Tags, tag)
		command.UpdatedAt = UnixTimeNow()
	}
}

func (command *Command) TagDelete(tag string) {
	if tag == "" {
		return
	}

	found := -1

	for i, t := range command.Tags {
		if t == tag {
			found = i
			break
		}
	}

	if found != -1 {
		command.Tags = append(command.Tags[:found], command.Tags[found+1:]...)
		command.UpdatedAt = UnixTimeNow()
	}
}

func (command *Command) Matches(criteria string) bool {
	criteria = strings.ToLower(criteria)

	matches := strings.Contains(strings.ToLower(command.Label), criteria)
	matches = matches || strings.Contains(strings.ToLower(command.Description), criteria)
	matches = matches || strings.Contains(strings.ToLower(command.Code), criteria)

	return matches
}

func (command *Command) Tagged(tag string) bool {
	for _, t := range command.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
