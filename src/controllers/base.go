package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"paxful/src/utils"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
)

type Server struct {
	DB           *sql.DB
	Router       *mux.Router
	RedisDB      *redis.Client
	PaxfulClient *http.Client
}

func (s *Server) Initialize() {
	//init logger
	utils.InitLogger()
	var err error

	//init DB
	var DSN = os.Getenv("db_user") + ":" + os.Getenv("db_pass") + "@tcp(" + os.Getenv("db_host") + ":" + os.Getenv("db_port") + ")/" + os.Getenv("db_name")
	s.DB, err = sql.Open("mysql", DSN)
	if err != nil {
		log.Println("Unable to connect to db:", err)
		os.Exit(3)
	}
	log.Println("Connected to db successfully")
	s.DB.SetMaxOpenConns(100)
	s.DB.SetMaxIdleConns(64)
	s.DB.SetConnMaxIdleTime(40)

	//init redis
	//init redis
	s.RedisDB = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	pong, err := s.RedisDB.Ping().Result()
	log.Println(fmt.Println(pong, err))

	//init Http CLient
	//Get token
	config := clientcredentials.Config{
		ClientID:     os.Getenv("PAXFUL_VILLAGERS_APP_ID"),
		ClientSecret: os.Getenv("PAXFUL_VILLAGERS_SECRET"),
		TokenURL:     os.Getenv("PAXFUL_ACCESS_TOKEN_URL"),
		Scopes:       []string{},
	}
	//setup context
	s.PaxfulClient = config.Client(context.Background())

}

func IntDigitsCount(number int) int {
	count := 0
	for number != 0 {
		number /= 10
		count += 1
	}
	return count

}
