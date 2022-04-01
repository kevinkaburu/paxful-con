package paxfulMs

import (
	"paxful/src/controllers"
)

var server = controllers.Server{}

//initialise server

func Run() {
	server.Initialize()

	server.PaxfulFetchOffers()

}
