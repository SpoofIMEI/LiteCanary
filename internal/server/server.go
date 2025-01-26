package server

import (
	"LiteCanary/internal/server/commands"
	"log"

	"github.com/gin-gonic/gin"
)

type Options struct {
	Debug            bool   `mapstructure:"debug"`
	DatabaseLocation string `mapstructure:"databaselocation"`
	Listener         string `mapstructure:"listener"`
	BasePath         string `mapstructure:"basepath"`
	NoRegistration   bool   `mapstructure:"noregistration"`
	PublicKey        string `mapstructure:"publickey"`
	PrivateKey       string `mapstructure:"privatekey"`
	Commander        *commands.Commander
}

type Server struct {
	Opts *Options
}

var (
	localOpts *Server
)

func New(opts *Options) *Server {
	localOpts = &Server{
		Opts: opts,
	}
	return localOpts
}

func (s *Server) StartApi() error {
	if !s.Opts.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.POST(s.Opts.BasePath+"/login", Login)
	router.POST(s.Opts.BasePath+"/reset", ResetPassword)
	router.POST(s.Opts.BasePath+"/register", Register)
	router.POST(s.Opts.BasePath+"/canary/new", NewCanary)
	router.POST(s.Opts.BasePath+"/canary/:id/wipe", WipeCanary)
	router.POST(s.Opts.BasePath+"/canary/update", UpdateCanary)
	router.GET(s.Opts.BasePath+"/canary", GetCanaries)
	router.GET(s.Opts.BasePath+"/trigger/:id", TriggerCanary)
	router.DELETE(s.Opts.BasePath+"/canary/:id", DeleteCanary)
	router.DELETE(s.Opts.BasePath+"/user", DeleteUser)

	log.Printf("running LiteCanary on %s", s.Opts.Listener)

	if localOpts.Opts.PublicKey != "" {
		return router.RunTLS(s.Opts.Listener, localOpts.Opts.PublicKey, localOpts.Opts.PrivateKey)
	}
	return router.Run(s.Opts.Listener)
}
