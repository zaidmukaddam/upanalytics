package project_test

import (
	"errors"
	"testing"

	"upanalytics/internal/project"
)

const (
	gid        = 1
	guid       = 1
	projectURL = "https://example.com"
	urlHost    = "example.com"
	urlScheme  = "https"
)

type storage struct{}

func (s *storage) SaveProject(project *project.Project, userId int) {}
func (s *storage) DeleteProject(project *project.Project)           {}
func (s *storage) UpdateProject(p *project.Project) error {
	return nil
}
func (s *storage) FindProjectById(id, uid int) (project.Project, error) {
	p := project.Project{}

	if id != gid || uid != guid {
		return p, errors.New("Project does not exist")
	}

	p.URL = projectURL

	return p, nil
}

var service = project.NewService(&storage{})

func TestFindProjectById(t *testing.T) {
	p, err := service.FindProject(gid, guid)
	if err != nil {
		t.Error(err)
	}

	if p.URL != projectURL {
		t.Errorf("p.URL: %s != %s", p.URL, projectURL)
	}

	if p.Host != urlHost {
		t.Errorf("p.Host: %s != %s", p.Host, urlHost)
	}

	p, err = service.FindProject(0, 0)
	if err == nil {
		t.Error("TestFindProjectById: should return err")
	}
}

func TestSaveProject(t *testing.T) {
	// Valid URL
	err := service.SaveProject(&project.Project{URL: projectURL}, guid)
	if err != nil {
		t.Error("TestSaveProject: should not return error")
	}

	// Not valid URL
	err = service.SaveProject(&project.Project{URL: "...."}, guid)
	if err == nil {
		t.Error("TestSaveProject: invalid URL should return error")
	}

	// Not supported scheme
	err = service.SaveProject(&project.Project{URL: "ftp://example.org"}, guid)
	if err == nil {
		t.Error("TestSaveProject: not supported scheme should return error")
	}
}
