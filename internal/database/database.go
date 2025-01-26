package database

import (
	"LiteCanary/internal/models"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Opts struct {
	Location       string
	Debug          bool
	NoRegistration bool
}

type CanaryDatabase struct {
	db *gorm.DB
}

// Creates a new canary databse object on a high level
func New(opts *Opts) (*CanaryDatabase, error) {
	config := &gorm.Config{}
	if !opts.Debug {
		config.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(sqlite.Open(opts.Location), config)
	if err != nil {
		return nil, err
	}
	canaryDatabase := &CanaryDatabase{
		db,
	}
	err = canaryDatabase.init()
	if err != nil {
		return nil, err
	}

	return canaryDatabase, nil
}

// > > LOWER LEVEL IMPLMENETATIONS < <

// Delete operations
func (canaryDatabase *CanaryDatabase) WipeCanary(id string) error {
	return canaryDatabase.db.Unscoped().Where("canaryid = ?", id).Delete(&models.TriggerEvent{}).Error
}
func (canaryDatabase *CanaryDatabase) DeleteCanary(id string) error {
	return canaryDatabase.db.Unscoped().Where("id = ?", id).Delete(&models.Canary{}).Error
}
func (canaryDatabase *CanaryDatabase) DeleteUser(username string) error {
	canaries, err := canaryDatabase.GetCanariesByOwner(username)
	if err != nil {
		return err
	}
	for _, canary := range canaries {
		err = canaryDatabase.DeleteCanary(canary.Id)
		if err != nil {
			return err
		}
	}
	err = canaryDatabase.db.Unscoped().Where("username = ?", username).Delete(&models.User{}).Error
	if err != nil {
		return err
	}
	return err
}
func (CanaryDatabase *CanaryDatabase) WipeExpiredTokens(expiration time.Duration) error {
	var Tokens []*models.Token
	CanaryDatabase.db.Find(&Tokens)
	for _, token := range Tokens {
		if time.Since(token.Lastactive) < expiration {
			continue
		}
		err := CanaryDatabase.db.Unscoped().Where("id = ?", token.Id).Delete(&models.Token{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// Read operations
func (canaryDatabase *CanaryDatabase) GetTriggerHistory(id string) (*[]models.TriggerEvent, error) {
	var history []models.TriggerEvent

	if err := canaryDatabase.db.Find(&history, "canaryid = ?", id).Error; err != nil {
		return nil, err
	}
	return &history, nil
}
func (canaryDatabase *CanaryDatabase) GetUserByToken(token string) (*models.User, error) {
	var user models.Token
	return user.User, canaryDatabase.db.First(&user, "id = ?", token).Error
}
func (canaryDatabase *CanaryDatabase) GetUser(username string) (models.User, error) {
	var user models.User
	return user, canaryDatabase.db.First(&user, "username = ?", username).Error
}
func (canaryDatabase *CanaryDatabase) GetCanariesByOwner(username string) ([]models.Canary, error) {
	var canaries []models.Canary
	err := canaryDatabase.db.Find(&canaries, "username = ?", username).Error
	return canaries, err
}
func (canaryDatabase *CanaryDatabase) GetCanariesByName(name string) ([]models.Canary, error) {
	var canaries []models.Canary
	err := canaryDatabase.db.Find(&canaries, "name = ?", name).Error
	return canaries, err
}
func (canaryDatabase *CanaryDatabase) GetCanaryByID(id string) (models.Canary, error) {
	var canary models.Canary
	err := canaryDatabase.db.First(&canary, "id = ?", id).Error
	return canary, err
}
func (CanaryDatabase *CanaryDatabase) ValidToken(token string) bool {
	var ses models.Token
	return CanaryDatabase.db.Where("id = ?", token).First(&ses).Error == nil
}

// Update operations
func (CanaryDatabase *CanaryDatabase) ResetPassword(username, newPassword string) error {
	user, err := CanaryDatabase.GetUser(username)
	if err != nil {
		return err
	}
	user.Password = newPassword
	CanaryDatabase.db.Save(&user)
	return nil
}
func (canaryDatabase *CanaryDatabase) UpdateCanary(newCanary *models.Canary) error {
	var canary models.Canary
	err := canaryDatabase.db.First(&canary, "id = ?", newCanary.Id).Error
	if err != nil {
		return err
	}
	canary.Name = newCanary.Name
	canary.Redirect = newCanary.Redirect
	canary.Type = newCanary.Type
	return canaryDatabase.db.Save(&canary).Error
}

// Create operations
func (canaryDatabase *CanaryDatabase) TriggerCanary(triggerEvent *models.TriggerEvent) error {
	return canaryDatabase.db.Create(triggerEvent).Error
}

func (canaryDatabse *CanaryDatabase) AddToken(user *models.User, Token string) error {
	return canaryDatabse.db.Create(&models.Token{
		Lastactive: time.Now(),
		Id:         Token,
		User:       user,
	}).Error
}
func (canaryDatabase *CanaryDatabase) AddUser(user *models.User) error {
	return canaryDatabase.db.Create(user).Error
}
func (canaryDatabase *CanaryDatabase) AddCanary(canary *models.Canary) error {
	return canaryDatabase.db.Create(canary).Error
}
func (canaryDatabase *CanaryDatabase) Close() error {
	db, err := canaryDatabase.db.DB()
	if err != nil {
		return err
	}
	db.Close()
	canaryDatabase = nil
	return nil
}

// Creates migrations for LiteCanary
func (canaryDatabase *CanaryDatabase) init() error {
	err := canaryDatabase.db.AutoMigrate(&models.User{}, &models.Canary{}, &models.Token{}, &models.TriggerEvent{})
	if err != nil {
		return err
	}
	return nil
}
