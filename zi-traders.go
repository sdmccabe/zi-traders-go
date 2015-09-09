package main

// ZI Traders Model
// Adapted from Axtell (2009)
// Stefan McCabe

// This is a port of Rob Axtell's ZI Traders model to Go.
// Original reference for the ZI model:
// Gode and Sunder, QJE, 1993

import (
	"fmt"
	"github.com/grd/stat"
	"github.com/pkg/profile"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

//globals
var numBuyers int = 1000000
var numSellers int = 1000000
var maxBuyerValue int = 30
var maxSellerValue int = 30
var maxNumberOfTrades int = 100000000
var numThreads int = 1
var buyersPerThread int = numBuyers / numThreads
var sellersPerThread int = numSellers / numThreads
var tradesPerThread int = maxNumberOfTrades / numThreads

//debugging
var countTrades uint64
var countActualTrades uint64

type agent struct {
	buyerOrSeller bool // true is buyer, false is seller
	quantityHeld  int
	value         int
	price         int
}

func (a agent) String() string {
	return fmt.Sprintf("buyer: %t, held: %d, value: %d, price: %d\n", a.buyerOrSeller, a.quantityHeld, a.value, a.price)
}
func initializeAgents() ([]agent, []agent) {
	// Create two slices of agents, one representing buyers and the other sellers.

	b := make([]agent, numBuyers)
	s := make([]agent, numSellers)

	for i := 0; i < numBuyers; i++ {
		b[i] = agent{
			buyerOrSeller: true,
			quantityHeld:  0,
			//value:         1 + rand.Intn(maxBuyerValue)}
			value: (rand.Int() % maxBuyerValue) + 1}
	}

	for i := 0; i < numSellers; i++ {
		s[i] = agent{
			buyerOrSeller: false,
			quantityHeld:  1,
			//value:         1 + rand.Intn(maxSellerValue)}
			value: (rand.Int() % maxBuyerValue) + 1}
	}

	return b, s
}

func openMarket(b []agent, s []agent) {
	// until we parallelize, this essentially just launches DoTrades() and computeStatistics()
	//for i := 0; i < maxNumberOfTrades; i++ {
	//doTrades(b, s, 1)
	//}
	var wg sync.WaitGroup
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(b []agent, s []agent, threadNum int) {
			defer wg.Done()
			//defer fmt.Printf("Finished thread number %d\n", threadNum)
			doTrades(b, s, threadNum)
			runtime.Gosched()
		}(b, s, i)
	}
	wg.Wait()
	//fmt.Printf("%v\n", b)
	fmt.Printf("%d out of %d possible trades executed (max: %d)\n", atomic.LoadUint64(&countActualTrades), atomic.LoadUint64(&countTrades), maxNumberOfTrades)
	computeStatistics(b, s)
}

func doTrades(b []agent, s []agent, threadNum int) {
	//Pair up buyers and sellers and execute trades if the bid and ask prices are compatible.
	//fmt.Println(threadNum)
	//fmt.Printf(" %d %d\n", bidPrice, askPrice)
	for i := 1; i < tradesPerThread; i++ { //why i=1?

		//bound the slice based on thread number
		lowerBuyerBound := threadNum * buyersPerThread
		upperBuyerBound := (threadNum+1)*buyersPerThread - 1
		lowerSellerBound := threadNum * sellersPerThread
		upperSellerBound := (threadNum+1)*sellersPerThread - 1

		//select buyer and seller
		buyerIndex := lowerBuyerBound + rand.Intn(upperBuyerBound-lowerBuyerBound)
		sellerIndex := lowerSellerBound + rand.Intn(upperSellerBound-lowerSellerBound)
		//fmt.Printf("buyerIndex: %d, sellerIndex: %d\n", buyerIndex, sellerIndex)

		//set bid and ask prices
		bidPrice := rand.Intn(b[buyerIndex].value) + 1
		askPrice := s[sellerIndex].value + rand.Intn(maxSellerValue-s[sellerIndex].value+1)

		//old bid/ask
		//bidPrice := (rand.Int() % b[buyerIndex].value) + 1
		//askPrice := s[sellerIndex].value + (rand.Int() % (maxSellerValue - s[sellerIndex].value + 1))
		var transactionPrice int
		//is a deal possible?
		if b[buyerIndex].quantityHeld == 0 && s[sellerIndex].quantityHeld == 1 && bidPrice >= askPrice {
			atomic.AddUint64(&countActualTrades, 1)
			// set transaction price
			transactionPrice = askPrice + rand.Intn(bidPrice-askPrice+1)
			b[buyerIndex].price = transactionPrice
			s[sellerIndex].price = transactionPrice

			// execute trade
			b[buyerIndex].quantityHeld = 1
			s[sellerIndex].quantityHeld = 0
		}

		atomic.AddUint64(&countTrades, 1)
	}
}

func computeStatistics(b []agent, s []agent) {
	// Compute some statistics for the run and output to STDOUT.
	numberBought := 0
	numberSold := 0
	sum := make(stat.IntSlice, 1)

	for _, x := range b {
		if x.quantityHeld == 1 {
			numberBought++
			sum = append(sum, int64(x.price))
		}
	}
	for _, x := range s {
		if x.quantityHeld == 0 {
			numberSold++
			sum = append(sum, int64(x.price))
		}
	}
	fmt.Printf("%d items bought and %d items sold\n", numberBought, numberSold)
	fmt.Printf("The average price = %f and the s.d. is %f\n", stat.Mean(sum), stat.Sd(sum))
}

func main() {
	defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	runtime.GOMAXPROCS(4)
	// seed RNG
	rand.Seed(time.Now().UTC().UnixNano())

	buyers, sellers := initializeAgents()
	openMarket(buyers, sellers)
}
