package main

import (
	"log"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	"gopkg.in/ini.v1"
)

// @title Rubus API
// @version 1.0
// @description Rubus REST API
// @termOfService http://swagger.io/terms/

// @contact.name Quentin Vaucher
// @contact.email quentin.vaucher3@master.hes-so.ch

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:1323

// @securityDefinitions.apikey jwt
// @in header
// @name Authorization

// @tag.name admin
// @tag.description Operations which require administrative rights
// @tag.name device
// @tag.description Operations about devices, such as provisioning or deployment
// @tag.name user
// @tag.description Operations about Users

type server struct {
	e   *echo.Echo
	db  *pg.DB
	cfg *ini.File
}

func main() {
	s := server{}

	s.cfg, _ = ini.Load("./conf/config.ini")

	// init db
	dbCfg := s.cfg.Section("database")
	s.db = pg.Connect(&pg.Options{
		Addr:     dbCfg.Key("address").String(),
		User:     dbCfg.Key("user").String(),
		Password: dbCfg.Key("password").String(),
		Database: dbCfg.Key("db_name").String(),
	})
	defer s.db.Close()

	if recreate, _ := dbCfg.Key("recreate").Bool(); recreate {
		log.Printf("Recreate database schemas")
		if err := deleteSchema(s.db); err != nil {
			panic(err)
		}
		if err := createSchema(s.db); err != nil {
			panic(err)
		}
		if err := createAdmin(s); err != nil {
			panic(err)
		}
	}

	// init REST API
	s.e = echo.New()
	createRESTEndpoints(s)

	s.e.Logger.Fatal(s.e.Start(":1323"))
}
