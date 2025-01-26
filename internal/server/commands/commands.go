package commands

import (
	"LiteCanary/internal/database"
	"LiteCanary/internal/errors"
	"LiteCanary/internal/models"
	"context"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Commander struct {
	canaryDatabase *database.CanaryDatabase
	secret         []byte
	expiration     time.Duration
	noRegistration bool
	ctx            context.Context
	cancel         context.CancelFunc
	wg             *sync.WaitGroup
}

func New(opts *database.Opts) (*Commander, error) {
	randBuffer := make([]byte, 1024)
	rand.Read(randBuffer)

	canaryDatabase, err := database.New(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	commander := &Commander{
		canaryDatabase: canaryDatabase,
		secret:         randBuffer,
		expiration:     1 * time.Hour,
		ctx:            ctx,
		noRegistration: opts.NoRegistration,
		cancel:         cancel,
		wg:             &sync.WaitGroup{},
	}

	commander.wg.Add(1)
	go commander.TokenWiper()

	return commander, nil
}

func (commander *Commander) Close() error {
	clear(commander.secret)
	commander.cancel()
	commander.wg.Wait()
	return commander.canaryDatabase.Close()
}

func (commander *Commander) TokenWiper() {
	var exit bool
	for !exit {
		time.Sleep(2 * time.Second)
		commander.canaryDatabase.WipeExpiredTokens(commander.expiration)
		select {
		case <-commander.ctx.Done():
			exit = true
		default:
			continue
		}
	}
	commander.wg.Done()
}

// Read operations
func (commander *Commander) ValidCanary(id string) bool {
	_, err := commander.canaryDatabase.GetCanaryByID(id)
	return err == nil
}
func (commander *Commander) GetTriggerHistory(token, id string) (*[]models.TriggerEvent, error) {
	if !commander.hasCanaryAccess(token, id) {
		return nil, errors.ErrNotAllowed
	}
	return commander.canaryDatabase.GetTriggerHistory(id)
}
func (commander *Commander) ValidToken(token string) bool {
	return commander.canaryDatabase.ValidToken(token)
}
func (commander *Commander) GetCanaries(token string) ([]models.Canary, error) {
	user, err := commander.GetUserByToken(token)
	if err != nil {
		return nil, err
	}
	canaries, err := commander.canaryDatabase.GetCanariesByOwner(user.Username)
	return canaries, err
}
func (commander *Commander) GetCanary(id string) (*models.Canary, error) {
	canary, err := commander.canaryDatabase.GetCanaryByID(id)
	if err != nil {
		return nil, err
	}
	return &canary, err
}
func (commander *Commander) GetUserByToken(token string) (*models.User, error) {
	user, err := commander.canaryDatabase.GetUserByToken(token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}
	return user, nil
}

// Delete operations
func (commander *Commander) DeleteUser(token string) error {
	if !commander.ValidToken(token) {
		return errors.ErrInvalidToken
	}
	user, err := commander.GetUserByToken(token)
	if err != nil {
		return err
	}
	return commander.canaryDatabase.DeleteUser(user.Username)
}
func (commander *Commander) WipeCanary(token, id string) error {
	if !commander.ValidToken(token) {
		return errors.ErrInvalidToken
	}
	if !commander.hasCanaryAccess(token, id) {
		return errors.ErrNotAllowed
	}
	return commander.canaryDatabase.WipeCanary(id)
}
func (commander *Commander) DeleteCanaryByID(token, id string) error {
	if !commander.ValidToken(token) {
		return errors.ErrInvalidToken
	}
	if !commander.hasCanaryAccess(token, id) {
		return errors.ErrNotAllowed
	}
	return commander.canaryDatabase.DeleteCanary(id)
}
func (commander *Commander) DeleteCanaryByName(token, name string) error {
	user, err := commander.GetUserByToken(token)
	if err != nil {
		return err
	}
	canaries, err := commander.canaryDatabase.GetCanariesByOwner(user.Username)
	if err != nil {
		return err
	}
	for _, canary := range canaries {
		if canary.Name != name {
			continue
		}
		err = commander.canaryDatabase.DeleteCanary(canary.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update operations
func (commander *Commander) ResetPassword(token, newPassword string) error {
	user, err := commander.GetUserByToken(token)
	if err != nil {
		return errors.ErrInvalidToken
	}
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	return commander.canaryDatabase.ResetPassword(user.Username, hashedPassword)
}

func (commander *Commander) UpdateCanary(token, id, name, canaryType, redirect string) error {
	user, err := commander.GetUserByToken(token)
	if err != nil {
		return errors.ErrInvalidToken
	}
	if !commander.hasCanaryAccess(token, id) {
		return errors.ErrNotAllowed
	}
	if !commander.ValidCanary(id) {
		return errors.ErrNotFound
	}
	canary := &models.Canary{
		Name:     name,
		Id:       id,
		Type:     canaryType,
		Redirect: redirect,
		User:     user,
	}
	err = commander.canaryDatabase.UpdateCanary(canary)
	if err != nil {
		return err
	}
	return nil
}

// Create operations
func (commander *Commander) TriggerCanary(triggerEvent *models.TriggerEvent) error {
	return commander.canaryDatabase.TriggerCanary(triggerEvent)
}
func (commander *Commander) AddCanary(token, name, canaryType, redirect string) (*models.Canary, error) {
	user, err := commander.GetUserByToken(token)
	if err != nil {
		return nil, err
	}
	canary := &models.Canary{
		Name:     name,
		Id:       uuid.NewString(),
		Type:     canaryType,
		Redirect: redirect,
		User:     user,
	}
	err = commander.canaryDatabase.AddCanary(canary)
	if err != nil {
		return nil, err
	}
	return canary, nil
}
func (commander *Commander) AddUser(username, password string) error {
	if commander.noRegistration {
		return errors.ErrRegistrationDisabled
	}
	_, err := commander.canaryDatabase.GetUser(username)
	if err == nil {
		return errors.ErrUsernameAlreadyRegistered
	}
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	err = commander.canaryDatabase.AddUser(&models.User{
		Username: username,
		Password: hashedPassword,
	})
	return err
}

// Helpers
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func (commander *Commander) hasCanaryAccess(token, id string) bool {
	user, err := commander.GetUserByToken(token)
	if err != nil {
		return false
	}
	canary, err := commander.GetCanary(id)
	if err != nil {
		return false
	}
	if canary.User.Username != user.Username {
		return false
	}
	return true
}

func (commander *Commander) Login(username, password string) (string, error) {
	user, err := commander.canaryDatabase.GetUser(username)
	if err != nil {
		return "", errors.ErrInvalidCredentials
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.ErrInvalidCredentials
	}
	return commander.newToken(&user), nil
}

func (commander *Commander) newToken(user *models.User) string {
	Token := fmt.Sprintf("%x", sha512.Sum512(append(commander.secret, []byte(uuid.NewString()+user.Username)...)))
	commander.canaryDatabase.AddToken(user, Token)
	return Token
}
