package wrapper

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/bennycio/bundle/api"
	"google.golang.org/grpc"
)

func UpdateUserApi(username string, updatedUser *api.User) error {

	u, err := url.Parse(ApiServerHost + "/users")
	if err != nil {
		return err
	}

	updatedBs, err := json.Marshal(updatedUser)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	buf.Write(updatedBs)

	req, err := http.NewRequest(http.MethodPatch, u.String(), buf)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func GetUserApi(username string, email string) (*api.User, error) {

	u, err := url.Parse(ApiServerHost + "/users")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("username", username)
	q.Set("email", email)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &api.User{}

	err = json.Unmarshal(bs, result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func InsertUser(user *api.User) error {
	conn, err := grpc.Dial(grpcAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := api.NewUsersServiceClient(conn)
	client.InsertUser(context.Background(), user)
	return nil
}

func GetUser(req *api.GetUserRequest) (*api.User, error) {
	conn, err := grpc.Dial(grpcAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := api.NewUsersServiceClient(conn)
	user, err := client.GetUser(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return user, nil
}