package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"github.com/bennycio/bundle/api"
	"github.com/bennycio/bundle/internal"
)

type repoService interface {
	DownloadPlugin(plugin *api.Plugin) ([]byte, error)
	UploadPlugin(user *api.User, plugin *api.Plugin, data io.Reader) error
	UploadThumbnail(user *api.User, plugin *api.Plugin, data io.Reader) error
}

type repoServiceImpl struct {
	Host string
	Port string
}

func NewRepoService(host string, port string) repoService {
	if host == "" {
		host = os.Getenv("REPO_HOST")
	}
	if port == "" {
		port = os.Getenv("REPO_PORT")
	}
	return &repoServiceImpl{
		Host: host,
		Port: port,
	}
}

func (r *repoServiceImpl) DownloadPlugin(plugin *api.Plugin) ([]byte, error) {

	scheme := "https://"
	u, err := url.Parse(fmt.Sprintf("%s%s:%s/repo/plugins", scheme, r.Host, r.Port))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Add("id", plugin.Id)
	q.Add("author", plugin.Author.Id)
	q.Add("version", plugin.Version)
	u.RawQuery = q.Encode()

	client := internal.NewTlsClient()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	access, err := getAccessToken()
	if err != nil {
		return nil, err
	}
	req.Header.Add("authorization", "Bearer "+access)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (r *repoServiceImpl) UploadPlugin(user *api.User, plugin *api.Plugin, data io.Reader) error {

	scheme := "https://"

	u, err := url.Parse(fmt.Sprintf("%s%s:%s/repo/plugins", scheme, r.Host, r.Port))
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if plugin.Id == "" || plugin.Version == "" || plugin.Author == nil {
		return errors.New("missing required fields")
	}

	writer.WriteField("id", plugin.Id)
	writer.WriteField("version", plugin.Version)
	writer.WriteField("author", plugin.Author.Id)

	part, err := writer.CreateFormFile("plugin", plugin.Id)
	if err != nil {
		return err
	}
	bs, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	_, err = part.Write(bs)
	if err != nil {
		return err
	}

	err = writer.Close()

	if err != nil {
		return err
	}

	client := internal.NewTlsClient()

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return err
	}
	access, err := getAccessToken()
	if err != nil {
		fmt.Println(err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	defer resp.Body.Close()
	return nil
}

func (r *repoServiceImpl) UploadThumbnail(user *api.User, plugin *api.Plugin, data io.Reader) error {
	fmt.Println("OJWO")
	scheme := "https://"

	u, err := url.Parse(fmt.Sprintf("%s%s:%s/repo/thumbnails", scheme, r.Host, r.Port))
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fmt.Println("OJWO")
	if plugin != nil {
		if plugin.Id == "" || plugin.Author == nil {
			return errors.New("specify plugin ID and author")
		}
		if plugin.Author.Id == "" {
			return errors.New("specify author ID")
		}
		writer.WriteField("author", plugin.Author.Id)
		writer.WriteField("plugin", plugin.Id)
	} else if user != nil {
		if user.Id == "" {
			return errors.New("specify a user id to upload a thumbnail to")
		}
		writer.WriteField("user", user.Id)
	} else {
		return errors.New("specify a user or a plugin to upload a thumbnail")
	}

	fmt.Println("OJWO")
	part, err := writer.CreateFormFile("thumbnail", "THUMBNAIL.webp")
	if err != nil {
		return err
	}
	bs, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	_, err = part.Write(bs)
	if err != nil {
		return err
	}
	err = writer.Close()

	if err != nil {
		return err
	}

	client := internal.NewTlsClient()
	fmt.Println("OJWO")
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return err
	}
	fmt.Println("OJWO")
	access, err := getAccessToken()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("OJWO")

	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println("OJWO")

	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	defer resp.Body.Close()
	return nil
}
