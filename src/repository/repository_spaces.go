package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dplabs/cbox/src/models"
	"github.com/dplabs/cbox/src/tools"
)

const (
	filenameSeparatorUser         = ":"
	filenameSeparatorOrganization = "="

	pathSpaces = "spaces"
)

func (repo *Repository) LoadSpaces() ([]*models.Space, bool) {

	isNewRepository := repo.initializeSpacesDirectory()

	spaces := []*models.Space{}

	files, err := ioutil.ReadDir(repo.resolve(pathSpaces))
	if err != nil {
		log.Fatalf("repository: could not read spaces: %v", err)
	}
	for _, f := range files {
		filename := f.Name()
		extension := filepath.Ext(filename)
		if extension == ".json" {
			namespaceType := models.TypeNone
			namespace := ""
			label := filename[0 : len(filename)-len(extension)]
			if strings.Contains(label, filenameSeparatorUser) {
				parts := strings.Split(label, filenameSeparatorUser)
				namespaceType = models.TypeUser
				namespace = parts[0]
				label = parts[1]
			} else if strings.Contains(label, filenameSeparatorOrganization) {
				parts := strings.Split(label, filenameSeparatorOrganization)
				namespaceType = models.TypeOrganization
				namespace = parts[0]
				label = parts[1]
			}
			spaces = append(spaces, repo.spaceLoadFile(namespaceType, namespace, label))
		}
	}
	return spaces, isNewRepository
}

func (repo *Repository) initializeSpacesDirectory() bool {
	spacesPath := repo.resolve(pathSpaces)
	return tools.CreateDirectoryIfNotExists(spacesPath)
}

func (repo *Repository) spaceLoadFile(namespaceType int, namespace string, label string) *models.Space {
	spacePath := repo.resolveSpaceFile(namespaceType, namespace, label)

	raw, err := ioutil.ReadFile(spacePath)
	if err != nil {
		log.Fatalf("repository: load space '%s-%s': could not read file '%s': %v", namespace, label, spacePath, err)
	}

	var space models.Space
	err = json.Unmarshal(raw, &space)

	if err != nil {
		log.Fatalf("repository: load space '%s-%s': could not parse JSON file: %v", namespace, label, err)
	}

	space.Selector, err = models.ParseSelector(space.ID)
	if err != nil {
		log.Fatalf("repository: load space '%s': space's ID is not a valid selector: %v", space.ID, err)
	}

	if space.Entries == nil {
		space.Entries = []*models.Command{}
	}

	for _, command := range space.Entries {
		selector, err := models.ParseSelectorMandatoryItem(command.ID)
		if err != nil {
			log.Fatalf("repository: load space '%s': command's ID (%s) is not a valid selector: %v", space.ID, command.ID, err)
		}
		command.Selector = selector
	}

	return &space
}

func (repo *Repository) resolveSpaceFile(namespaceType int, namespace string, label string) string {
	filename := label
	if namespaceType != models.TypeNone {
		separator := filenameSeparatorUser
		if namespaceType == models.TypeOrganization {
			separator = filenameSeparatorOrganization
		}
		filename = fmt.Sprintf("%s%s%s", namespace, separator, label)
	}
	filename = filename + ".json"
	return repo.resolve(pathSpaces, filename)
}

func (repo *Repository) Persist(space *models.Space) {

	space.ID = space.Selector.String()

	for _, command := range space.Entries {
		command.Selector.NamespaceType = space.Selector.NamespaceType
		command.Selector.Namespace = space.Selector.Namespace
		command.Selector.Space = space.Label

		command.ID = command.Selector.String()
	}

	raw, err := json.MarshalIndent(space, "", "  ")
	if err != nil {
		log.Fatalf("repository: store space '%s': could not generate JSON: %v", space.String(), err)
	}

	file := repo.resolveSpaceFile(space.Selector.NamespaceType, space.Selector.Namespace, space.Label)
	err = ioutil.WriteFile(file, raw, 0644)
	if err != nil {
		log.Fatalf("repository: store space '%s': could not write JSON file (%s): %v", space.String(), file, err)
	}
}

func (repo *Repository) Delete(selector *models.Selector) {
	file := repo.resolveSpaceFile(selector.NamespaceType, selector.Namespace, selector.Space)
	err := os.Remove(file)
	if err != nil {
		log.Fatalf("repository: delete space '%s': %v", selector.String(), err)
	}
}
