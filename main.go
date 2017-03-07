package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	dockerURI = flag.String("docker-uri", "unix:///var/run/docker.sock", "The docker URI.")
	input     = flag.String("input", "", "The file to read the JSON from.")
)

// ImageInfo contains the information about each image that needs to be included
// in the manifest.
type ImageInfo struct {
	RepoTag string `json:"repo-tag"`
	ImageID string `json:"image-id"`
	GitRef  string `json:"git-ref"`
}

// InputMap is the top-level struct that the JSON input is parsed into
type InputMap struct {
	Hostname string      `json:"hostname"`
	Date     string      `json:"date"`
	Images   []ImageInfo `json:"images"`
}

// NewInputMap reads the contents of the file at 'path' and parses it into a new
// *InputMap.
func NewInputMap(path string) (*InputMap, error) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	inputMap := &InputMap{}
	err = json.Unmarshal(input, inputMap)
	return inputMap, err
}

func main() {
	flag.Parse()

	if *input == "" {
		log.Fatal("--input must be set")
	}

	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	d, err := client.NewClient(*dockerURI, "v1.22", nil, defaultHeaders)
	if err != nil {
		log.Fatalf("Error creating docker client: %s", err)
	}

	inputMap, err := NewInputMap(*input)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// grab a list of all of the local images.
	listedImages, err := d.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	valid := true
	for _, inputImage := range inputMap.Images {
		found := false
		for _, listedImage := range listedImages {
			for _, listedRepoTag := range listedImage.RepoTags {
				if listedRepoTag == inputImage.RepoTag {
					found = true
					if listedImage.ID != inputImage.ImageID {
						fmt.Printf(
							"Docker image ID does not match for %s: %s, %s\n",
							inputImage.RepoTag,
							listedImage.ID,
							inputImage.ImageID,
						)
						valid = false
					}
				}
			}
		}
		if !found {
			fmt.Printf("Docker image %s was not found\n", inputImage.RepoTag)
			valid = false
		}
	}

	if !valid {
		os.Exit(-1)
	}
}
