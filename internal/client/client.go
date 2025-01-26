package client

import (
	"LiteCanary/internal/errors"
	"LiteCanary/internal/models"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Client struct {
	LoggedIn bool
	Username string
	session  string
	url      string
	canaries []models.FilteredCanary
}

var (
	customHttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Disable SSL certificate verification
			},
		},
	}
)

func New(url string) *Client {
	var client Client
	client.url = url
	return &client
}

// Get info for a specific canary
func (client *Client) GetCanary(id string) (*models.FilteredCanary, error) {
	client.UpdateCanaries()
	for _, canary := range client.canaries {
		if canary.Id == id {
			return &canary, nil
		}
	}
	return nil, errors.ErrNotFound
}

// Get canaries
type canariesResp struct {
	Canaries []models.FilteredCanary
}

func (client *Client) UpdateCanaries() (*canariesResp, error) {
	req, err := http.NewRequest("GET", client.url+"/canary", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.session))
	resp, err := customHttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		return nil, errors.ErrInvalidToken
	}
	if resp.StatusCode != 200 {
		return nil, errors.ErrSomethingWentWrong
	}
	var respDecoded canariesResp
	err = json.NewDecoder(resp.Body).Decode(&respDecoded)
	if err != nil {
		return nil, err
	}
	client.canaries = respDecoded.Canaries
	return &respDecoded, nil
}

// delete canary
func (client *Client) DeleteUser() error {
	req, err := http.NewRequest("DELETE", client.url+"/user", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.session))
	resp, err := customHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidToken
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	client.LoggedIn = false
	client.session = ""
	return nil
}

// wipe canary
func (client *Client) WipeCanary(id string) error {
	req, err := http.NewRequest("POST", client.url+"/canary/"+id+"/wipe", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.session))
	resp, err := customHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidCredentials
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	return nil
}

// delete canary
func (client *Client) DeleteCanary(id string) error {
	req, err := http.NewRequest("DELETE", client.url+"/canary/"+id, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.session))
	resp, err := customHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidCredentials
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	return nil
}

// Update canary
type updateCanaryReq struct {
	Name     string
	Type     string
	Redirect string
	Id       string
}

func (client *Client) UpdateCanary(id, name, canaryType, redirect string) error {
	canaryReq := updateCanaryReq{
		Name:     name,
		Type:     canaryType,
		Redirect: redirect,
		Id:       id,
	}
	marshaled, err := json.Marshal(canaryReq)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", client.url+"/canary/update", bytes.NewReader(marshaled))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.session))
	resp, err := customHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidToken
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	return nil
}

// New canary
type newCanaryResp struct {
	Name      string
	Id        string
	CreatedAt string
}

type newCanaryReq struct {
	Name     string
	Type     string
	Redirect string
}

func (client *Client) NewCanary(name, canaryType string) error {
	var redirect string
	if canaryType == "redirect" {
		fmt.Print("URL to redirect to: ")
		reader := bufio.NewReader(os.Stdin)
		redirect, _ = reader.ReadString('\n')
	}
	canaryReq := newCanaryReq{
		Name:     name,
		Type:     canaryType,
		Redirect: redirect,
	}
	marshaled, err := json.Marshal(canaryReq)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", client.url+"/canary/new", bytes.NewReader(marshaled))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.session))
	resp, err := customHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidToken
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	var respDecoded newCanaryResp
	err = json.NewDecoder(resp.Body).Decode(&respDecoded)
	if err != nil {
		return err
	}
	return nil
}

// Login
type loginResp struct {
	Session string
}

type loginReq struct {
	Username string
	Password string
}

func (client *Client) Login(username, password string) error {
	marshaled, err := json.Marshal(loginReq{
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}
	resp, err := customHttpClient.Post(client.url+"/login", "application/json", bytes.NewReader(marshaled))
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidCredentials
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	var respDecoded loginResp
	err = json.NewDecoder(resp.Body).Decode(&respDecoded)
	if err != nil {
		return err
	}
	client.session = respDecoded.Session
	client.LoggedIn = true
	client.Username = username
	return nil
}

// Password reset
type resetReq struct {
	Password string
}

func (client *Client) ResetPassword(password string) error {
	marshaled, err := json.Marshal(resetReq{
		Password: password,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", client.url+"/reset", bytes.NewReader(marshaled))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.session))
	resp, err := customHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidToken
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	return nil
}

// Registration
type registerReq struct {
	Username string
	Password string
}

func (client *Client) Register(username, password string) error {
	marshaled, err := json.Marshal(registerReq{
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}
	resp, err := customHttpClient.Post(client.url+"/register", "application/json", bytes.NewReader(marshaled))
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.ErrInvalidCredentials
	}
	if resp.StatusCode != 200 {
		return errors.ErrSomethingWentWrong
	}
	return nil
}
