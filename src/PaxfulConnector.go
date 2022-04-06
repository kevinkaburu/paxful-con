package paxfulMs

import (
	"paxful/src/controllers"
	"sync"
)

var server = controllers.Server{}

//initialise server

func Run() {

	server.Initialize()
	var wg sync.WaitGroup

	wg.Add(2)
	go server.PaxfulFetchOffers(&wg)
	go server.Forex(&wg)
	go server.MCapExchange(&wg)
	wg.Wait()

}
