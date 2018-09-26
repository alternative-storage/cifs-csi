package cifs

import (
	"errors"
)

type volumeOptions struct {
	Server string `json:"server"`
	Share  string `json:"share"`
}

func extractOption(dest *string, optionLabel string, options map[string]string) error {
	if opt, ok := options[optionLabel]; !ok {
		return errors.New("Missing required field " + optionLabel)
	} else {
		*dest = opt
		return nil
	}
}

func newVolumeOptions(volOptions map[string]string) (*volumeOptions, error) {
	var (
		opts volumeOptions
		err  error
	)

	if err = extractOption(&opts.Server, "server", volOptions); err != nil {
		return nil, err
	}

	return &opts, nil
}
