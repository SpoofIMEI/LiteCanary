package tests

import (
	"LiteCanary/internal/client"
	"net/http"
	"testing"
)

var (
	url = "http://127.0.0.1:8080/api"
)

func TestClient(t *testing.T) {
	client := client.New(url)
	err := client.Register("peter", "test")
	if err != nil {
		t.Fatalf("couldn't register: %s", err.Error())
	}
	t.Log("SUCCESS -> register")
	err = client.Login("peter", "test")
	if err != nil {
		t.Fatalf("couldn't login: %s", err.Error())
	}
	t.Log("SUCCESS -> login")
	err = client.NewCanary("test", "image")
	if err != nil {
		t.Fatalf("couldn't create canary: %s", err.Error())
	}
	t.Log("SUCCESS -> new canary")
	canariesResp, err := client.UpdateCanaries()
	if err != nil {
		t.Fatalf("couldn't get canaries: %s", err.Error())
	}
	id := canariesResp.Canaries[0].Id
	t.Log("SUCCESS -> get canaries")
	resp, err := http.Get(url + "/trigger/" + id)
	if err != nil {
		t.Fatalf("invalid canary trigger resp: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Fatalf("invalid canary trigger resp status code: %d", resp.StatusCode)
	}
	t.Log("SUCCESS -> canary trigger")

	err = client.WipeCanary(id)
	if err != nil {
		t.Fatalf("couldn't wipe canary events: %s", err.Error())
	}
	canariesResp, err = client.UpdateCanaries()
	if err != nil {
		t.Fatalf("couldn't get canaries: %s", err.Error())
	}
	if canariesResp.Canaries[0].History != nil {
		t.Fatal("couldn't wipe canary events (changes not reflected)")
	}
	t.Log("SUCCESS -> canary wipe")

	err = client.ResetPassword("test1234")
	if err != nil {
		t.Fatalf("couldn't reset user password: %s", err.Error())
	}
	err = client.Login("peter", "test1234")
	if err != nil {
		t.Fatalf("couldn't reset user password (change didn't reflect): %s", err.Error())
	}
	t.Log("SUCCESS -> password reset")
	err = client.DeleteUser()
	if err != nil {
		t.Fatalf("couldn't remove user: %s", err.Error())
	}
	t.Log("SUCCESS -> user deleted")
}
