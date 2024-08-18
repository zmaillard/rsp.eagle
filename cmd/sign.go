package cmd

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

type SignRequest struct {
	Path       string   `json:"path"`
	Name       string   `json:"name"`
	Website    string   `json:"website"`
	Tags       []string `json:"tags"`
	Annotation string   `json:"annotation"`
	FolderId   string   `json:"folderId"`
}

type Sign struct {
	Id      int      `db:"id"`
	Title   string   `db:"title"`
	ImageId string   `db:"imageid"`
	Tags    []string `db:"tags"`
	State   string   `db:"state"`
}

func (s Sign) String() string {
	t := strings.Join(s.Tags, ",")
	return fmt.Sprintf("%v::%s::%s::%s::%s", s.Id, s.Title, s.ImageId, s.State, t)
}

func (s Sign) BuildPath(cfg *config) string {
	return filepath.Join(cfg.SignDirectory, s.ImageId, s.ImageId+".jpg")
}

func (s Sign) BuildWebsite(cfg *config) (string, error) {
	return url.JoinPath(cfg.BaseWebsite, "sign", s.ImageId)
}

func (s Sign) BuildRequest(cfg *config, folders map[string]string) SignRequest {
	website, err := s.BuildWebsite(cfg)
	if err != nil {
		panic(err)
	}
	return SignRequest{
		Path:       s.BuildPath(cfg),
		Name:       s.Title,
		Website:    website,
		Tags:       s.Tags,
		Annotation: s.ImageId,
		FolderId:   folders[s.State],
	}
}
