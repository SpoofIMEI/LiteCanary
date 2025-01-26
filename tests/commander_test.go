package tests

import (
	CanaryDatabase "LiteCanary/internal/database"
	"LiteCanary/internal/server/commands"
	"testing"
)

func TestCommander(t *testing.T) {
	// Normal usage
	commander, err := commands.New(&CanaryDatabase.Opts{
		Location: ":memory:",
	})
	if err != nil {
		t.Fatalf("couldn't create commander: %s", err.Error())
	}
	t.Log("SUCCESS -> commander created and initialized")

	err = commander.AddUser("peter", "password1234")
	if err != nil {
		t.Fatalf("couldn't create user: %s", err.Error())
	}
	t.Log("SUCCESS -> user created")

	token, err := commander.Login("peter", "password1234")
	if err != nil {
		t.Fatalf("couldn't log in: %s", err.Error())
	}
	t.Logf("SUCCESS -> login successful  token:%s", token)

	canary, err := commander.AddCanary(token, "testcanary", "text", "")
	if err != nil {
		t.Fatalf("couldn't create canary: %s", err.Error())
	}
	canaries, err := commander.GetCanaries(token)
	if err != nil {
		t.Fatalf("couldn't get canaries: %s", err.Error())
	}
	if len(canaries) != 1 {
		t.Fatal("invalid amount of canary tokens for user")
	}
	t.Log("SUCCESS -> canary creation for a user")

	err = commander.UpdateCanary(token, canary.Id, "testcanary2", "image", "")
	if err != nil {
		t.Fatalf("couldn't update canary: %s", err.Error())
	}
	canary, err = commander.GetCanary(canary.Id)
	if err != nil {
		t.Fatalf("couldn't get canary: %s", err.Error())
	}
	if canary.Name != "testcanary2" && canary.Type == "image" {
		t.Fatalf("canary update didn't reflect")
	}
	t.Log("SUCCESS -> canary update")

	commander.DeleteCanaryByName(token, "testcanary2")
	canaries, err = commander.GetCanaries(token)
	if err != nil {
		t.Fatalf("couldn't get canaries: %s", err.Error())
	}
	if len(canaries) == 1 {
		t.Fatal("canary deletation didn't work")
	}
	t.Log("SUCCESS -> canary deletation")

	err = commander.ResetPassword(token, "test1234")
	if err != nil {
		t.Fatalf("couldn't reset password: %s", err.Error())
	}
	_, err = commander.Login("peter", "test1234")
	if err != nil {
		t.Fatalf("couldn't reset password (can't login with new password): %s", err.Error())
	}
	t.Log("SUCCESS -> password reset")

	// token validation checks
	_, err = commander.AddCanary("DOESN'T EXIST", "testcanary", "text", "")
	if err == nil {
		t.Fatalf("could create a canary with non existant token")
	}
	_, err = commander.GetCanaries("DOESN'T EXIST")
	if err == nil {
		t.Fatalf("could retrieve canaries with non existant token")
	}
	err = commander.DeleteCanaryByName("DOESN'T EXIST", "testcanary")
	if err == nil {
		t.Fatalf("could delete a canary with non existant token")
	}
	t.Log("SUCCESS -> token validation")

	commander.Close()

	t.Log("> > all commander checks passed! < <")
}
