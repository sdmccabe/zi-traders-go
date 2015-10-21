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

type agent struct {
	buyerOrSeller bool // true is buyer, false is seller
	quantityHeld  int
	value         int
	price         int
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
			value:         rand.Intn(maxBuyerValue) + 1}
	}

	for i := 0; i < numSellers; i++ {
		s[i] = agent{
			buyerOrSeller: false,
			quantityHeld:  1,
			value:         rand.Intn(maxSellerValue) + 1}
	}

	return b, s
}

// Divide the agent population into chunks, have these chunks perform trades,
// then compute market statistics.
func openMarket() {
	var wg sync.WaitGroup

	if verbose {
		fmt.Println(buyers)
	}

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(threadNum int) {
			defer wg.Done()
			if verbose {
				defer fmt.Printf("Finished thread number %d\n", threadNum)
			}
			doTrades(threadNum)
		}(i)
	}
	wg.Wait() //block until all threads are done for safety

	if verbose {
		fmt.Println(buyers)
	}

	computeStatistics()
}

//Pair up buyers and sellers and execute trades if the bid and ask prices are compatible.
func doTrades(threadNum int) {
	// Each thread needs its own random source to prevent excessive blocking on rand.
	// Adding these lines sped the model up approx. 9 times.
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	for i := 1; i < tradesPerThread; i++ { //why i=1?

		//bound the slice based on thread number
		lowerBuyerBound := threadNum * buyersPerThread
		upperBuyerBound := (threadNum+1)*buyersPerThread - 1
		lowerSellerBound := threadNum * sellersPerThread
		upperSellerBound := (threadNum+1)*sellersPerThread - 1

		//select buyer and seller
		buyerIndex := lowerBuyerBound + generator.Intn(upperBuyerBound-lowerBuyerBound)
		sellerIndex := lowerSellerBound + generator.Intn(upperSellerBound-lowerSellerBound)

		//set bid and ask prices
		bidPrice := generator.Intn(buyers[buyerIndex].value) + 1
		askPrice := sellers[sellerIndex].value + generator.Intn(maxSellerValue-sellers[sellerIndex].value+1)

		var transactionPrice int

		//is a deal possible?
		if buyers[buyerIndex].quantityHeld == 0 && sellers[sellerIndex].quantityHeld == 1 && bidPrice >= askPrice {
			// set transaction price
			transactionPrice = askPrice + generator.Intn(bidPrice-askPrice+1)
			buyers[buyerIndex].price = transactionPrice
			sellers[sellerIndex].price = transactionPrice

			// execute trade
			buyers[buyerIndex].quantityHeld = 1
			sellers[sellerIndex].quantityHeld = 0
		}
	}
}

// Compute some statistics for the run and output to STDOUT.
func computeStatistics() {
	numberBought := 0
	numberSold := 0
	sum := make(stat.IntSlice, 0)

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
	flag.IntVar(&numThreads, "p", runtime.NumCPU()*2, "number of goroutine to use")
	flag.BoolVar(&verbose, "v", false, "verbose (track goroutines)")
	flag.BoolVar(&profiling, "profile", false, "enable CPU profiling")
	flag.Parse()

	if profiling {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	}

	buyersPerThread = numBuyers / numThreads
	sellersPerThread = numSellers / numThreads
	tradesPerThread = maxNumberOfTrades / numThreads

	// seed RNG
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Printf("numThreads: %d\n", numThreads)

	buyers, sellers = initializeAgents()
	openMarket()
}
