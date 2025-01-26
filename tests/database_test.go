package tests

import (
	CanaryDatabase "LiteCanary/internal/database"
	"LiteCanary/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDatabase(t *testing.T) {
	// Init
	dbOptions := CanaryDatabase.Opts{
		Location: ":memory:",
	}
	db, err := CanaryDatabase.New(&dbOptions)
	if err != nil {
		t.Fatalf("couldn't create database: %s", err.Error())
	}
	if db == nil {
		t.Fatal("db is nil")
	}
	t.Log("SUCCESS -> database initialized")

	// Create
	seedDummyData := func() {
		user := models.User{Username: "peter", Password: "test"}
		err = db.AddUser(&user)
		if err != nil {
			t.Fatalf("couldn't create user: %s", err.Error())
		}
		err = db.AddCanary(&models.Canary{
			Name: "test",
			Id:   uuid.NewString(),
			User: &user,
		})
		if err != nil {
			t.Fatalf("couldn't create canary: %s", err.Error())
		}
		t.Log("SUCCESS -> database seeded")
	}
	seedDummyData()

	// Delete user
	err = db.DeleteUser("peter")
	if err != nil {
		t.Fatalf("couldn't delete user: %s", err.Error())
	}
	ownersCanaries, err := db.GetCanariesByOwner("peter")
	if err != nil {
		t.Fatalf("couldn't get canaries for a user: %s", err.Error())
	}
	if len(ownersCanaries) != 0 {
		t.Fatalf("couldn't delete user (no actual effect): %s", err.Error())
	}
	t.Log("SUCCESS -> user deletation")

	seedDummyData()

	// Read
	canaries, err := db.GetCanariesByName("test")
	if err != nil {
		t.Fatalf("couldn't get canaries for a user: %s", err.Error())
	}
	id1 := canaries[0].Id
	ownersCanaries, err = db.GetCanariesByOwner("peter")
	if err != nil {
		t.Fatalf("couldn't get canaries for a user: %s", err.Error())
	}
	id2 := ownersCanaries[0].Id
	if id1 != id2 {
		t.Fatalf("UUID doesn't match (id1:%s id2:%s)", id1, id2)
	}
	t.Log("SUCCESS -> database read operation")

	// Delete canary
	err = db.DeleteCanary(id1)
	if err != nil {
		t.Fatalf("couldn't delete canary: %s", err.Error())
	}
	ownersCanaries, err = db.GetCanariesByOwner("peter")
	if err != nil {
		t.Fatalf("couldn't get canaries for a user: %s", err.Error())
	}
	if len(ownersCanaries) != 0 {
		t.Fatalf("couldn't delete canary (no actual effect): %s", err.Error())
	}
	t.Log("SUCCESS -> canary deletation")

	// Tokens
	err = db.AddToken(&models.User{
		Username: "peter",
		Password: "password1234",
	}, "testToken")
	if err != nil {
		t.Fatalf("couldn't create Token: %s", err.Error())
	}
	t.Log("SUCCESS -> Token creation")

	if !db.ValidToken("testToken") {
		t.Fatal("Token validation not working")
	}
	if db.ValidToken("doesn't exist") {
		t.Fatal("Token validation not working")
	}
	t.Log("SUCCESS -> Token validation")

	err = db.WipeExpiredTokens(time.Second * 30)
	if err != nil {
		t.Fatalf("Token wipe not working: %s", err.Error())
	}
	if !db.ValidToken("testToken") {
		t.Fatal("Token wipe not working (wipes too quickly)")
	}

	time.Sleep(time.Second * 2)
	err = db.WipeExpiredTokens(time.Second * 1)
	if err != nil {
		t.Fatalf("Token wipe not working: %s", err.Error())
	}
	if db.ValidToken("testToken") {
		t.Fatal("Token wipe not working (doesn't wipe)")
	}

	t.Log("> > all database checks passed! < <")
}
