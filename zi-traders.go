package main

// ZI Traders Model
// Adapted from Axtell (2009)
// Stefan McCabe

// This is a port of Rob Axtell's ZI Traders model to Go.
// Original reference for the ZI model:
// Gode and Sunder, QJE, 1993

import (
	"flag"
	"fmt"
	"github.com/grd/stat"
	"github.com/pkg/profile"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

//globals
var numBuyers = 1200000
var numSellers = 1200000

var maxBuyerValue = 30
var maxSellerValue = 30
var maxNumberOfTrades = 100000000
var numThreads int
var buyersPerThread int
var sellersPerThread int
var tradesPerThread int
var buyers []agent
var sellers []agent
var verbose bool
var profiling bool
var randomNumbers chan int

type agent struct {
	buyerOrSeller bool // true is buyer, false is seller
	quantityHeld  int
	value         int
	price         int
	lock          sync.Mutex
}

func (a agent) String() string {
	return fmt.Sprintf("buyer: %t, held: %d, value: %d, price: %d\n", a.buyerOrSeller, a.quantityHeld, a.value, a.price)
}

// Create two slices of agents, one representing buyers and the other sellers.
func initializeAgents() ([]agent, []agent) {

	b := make([]agent, numBuyers)
	s := make([]agent, numSellers)

	for i := 0; i < numBuyers; i++ {
		b[i] = agent{
			buyerOrSeller: true,
			quantityHeld:  0,
			value:         (rand.Int() % maxBuyerValue) + 1}
	}

	for i := 0; i < numSellers; i++ {
		s[i] = agent{
			buyerOrSeller: false,
			quantityHeld:  1,
			value:         (rand.Int() % maxBuyerValue) + 1}
	}

	return b, s
}

// Execute each potential trade on its own goroutine, blocking while needed, then
// compute statistics and output.
func openMarket() {
	var wg sync.WaitGroup
	wg.Add(maxNumberOfTrades)

	for i := 0; i < maxNumberOfTrades; i++ {
		buyer := &buyers[rand.Intn(numBuyers)]
		seller := &sellers[rand.Intn(numSellers)]
		go func(b, s *agent) {
			defer wg.Done()

			buyer.lock.Lock()
			seller.lock.Lock()
			makeTrade(b, s)
			buyer.lock.Unlock()
			seller.lock.Unlock()
		}(buyer, seller)
	}
	wg.Wait()
	computeStatistics()
}

// Execute a trade between buyer and seller if possible.
func makeTrade(buyer, seller *agent) {

	//set bid and ask prices
	//bidPrice := rand.Intn(buyer.value) + 1
	bidPrice := (<-randomNumbers % buyer.value) + 1
	//askPrice := seller.value + rand.Intn(maxSellerValue-seller.value+1)
	askPrice := seller.value + (<-randomNumbers % (maxSellerValue - seller.value + 1))

	var transactionPrice int

	//is a deal possible?
	if buyer.quantityHeld == 0 && seller.quantityHeld == 1 && bidPrice >= askPrice {
		// set transaction price
		//transactionPrice = askPrice + rand.Intn(bidPrice-askPrice+1)
		transactionPrice = askPrice + (<-randomNumbers % (bidPrice - askPrice + 1))
		buyer.price = transactionPrice
		seller.price = transactionPrice

		// execute trade
		buyer.quantityHeld = 1
		seller.quantityHeld = 0
	}
}

// Compute some statistics for the run and output to STDOUT.
func computeStatistics() {
	numberBought := 0
	numberSold := 0
	sum := make(stat.IntSlice, 1)

	for _, x := range buyers {
		if x.quantityHeld == 1 {
			numberBought++
			sum = append(sum, int64(x.price))
		}
	}
	for _, x := range sellers {
		if x.quantityHeld == 0 {
			numberSold++
			sum = append(sum, int64(x.price))
		}
	}

	fmt.Printf("%d items bought and %d items sold\n", numberBought, numberSold)
	fmt.Printf("The average price = %f and the s.d. is %f\n", stat.Mean(sum), stat.Sd(sum))
}

func main() {
	fmt.Printf("\nZERO INTELLIGENCE TRADERS\n")
	flag.IntVar(&numThreads, "p", runtime.NumCPU()*2, "number of goroutines to use")
	flag.BoolVar(&verbose, "v", false, "verbose (track goroutines)")
	flag.BoolVar(&profiling, "profile", false, "enable CPU profiling")
	flag.Parse()

	if profiling {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	}
	buyersPerThread = int(float64(numBuyers) / float64(numThreads))
	sellersPerThread = int(float64(numSellers) / float64(numThreads))
	tradesPerThread = int(float64(maxNumberOfTrades) / float64(numThreads))

	// seed RNG
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Printf("numThreads: %d\n", numThreads)
	randomNumbers = make(chan int, 10000000) // arbitary buffer size
	go func() {
		for {
			randomNumbers <- rand.Int()
		}
	}()
	buyers, sellers = initializeAgents()
	openMarket()
}
