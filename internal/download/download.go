package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

const (
	bucket = "flyem-male-cns"
	prefix = "v1.0/connectome-data/flat-connectome/"

	Annotations       = "body-annotations-male-cns-v1.0-minconf-0.5.feather"
	Neurotransmitters = "body-neurotransmitters-male-cns-v1.0.feather"
	Weights           = "connectome-weights-male-cns-v1.0-minconf-0.5.feather"
)

var Files = []string{Annotations, Neurotransmitters, Weights}

var client = &http.Client{}

func All(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, name := range Files {
		if err := fetch(dir, name); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func remoteSize(name string) (int64, error) {
	endpoint := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s?fields=size", bucket, url.PathEscape(prefix+name))
	resp, err := client.Get(endpoint)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("metadata request: %s", resp.Status)
	}
	var body struct {
		Size string `json:"size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	return strconv.ParseInt(body.Size, 10, 64)
}

func fetch(dir, name string) error {
	size, err := remoteSize(name)
	if err != nil {
		return err
	}
	dest := filepath.Join(dir, name)
	if st, err := os.Stat(dest); err == nil && st.Size() == size {
		fmt.Printf("ok   %s (already downloaded)\n", name)
		return nil
	}
	endpoint := fmt.Sprintf("https://storage.googleapis.com/%s/%s%s", bucket, prefix, name)
	resp, err := client.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download request: %s", resp.Status)
	}
	tmp, err := os.CreateTemp(dir, name+".*.part")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	written, err := io.Copy(tmp, resp.Body)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if written != size {
		return fmt.Errorf("downloaded %d bytes, expected %d", written, size)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return err
	}
	fmt.Printf("ok   %s (%d bytes)\n", name, written)
	return nil
}
