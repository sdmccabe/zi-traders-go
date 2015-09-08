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
	"math/rand"
	"time"
)

//globals
var numBuyers int = 1200000
var numSellers int = 1200000
var maxBuyerValue int = 30
var maxSellerValue int = 30
var maxNumberOfTrades int = 100000000

type agent struct {
	buyerOrSeller bool // true is buyer, false is seller
	quantityHeld  int
	value         int
	price         int
}

func initializeAgents() ([]agent, []agent) {
	// Create two slices of agents, one representing buyers and the other sellers.

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
			value:         (rand.Int() % maxSellerValue) + 1}

	}
	return b, s
}

func openMarket(b []agent, s []agent) {
	// until we parallelize, this essentially just launches DoTrades() and computeStatistics()
	for i := 0; i < maxNumberOfTrades; i++ {
		doTrades(b, s)
	}
	computeStatistics(b, s)
}

func doTrades(b []agent, s []agent) {
	//Pair up buyers and sellers and execute trades if the bid and ask prices are compatible.

	//select buyer and seller
	buyerIndex := rand.Intn(len(b))
	sellerIndex := rand.Intn(len(s))

	//set bid and ask prices
	bidPrice := (rand.Int() % b[buyerIndex].value) + 1
	askPrice := s[sellerIndex].value + (rand.Int() % (maxSellerValue - s[sellerIndex].value + 1))
	var transactionPrice int
	//fmt.Printf(" %d %d\n", bidPrice, askPrice)

	//is a deal possible?
	if b[buyerIndex].quantityHeld == 0 && s[sellerIndex].quantityHeld == 1 && bidPrice >= askPrice {
		// set transaction price
		transactionPrice = askPrice + rand.Int()%(bidPrice-askPrice+1)
		b[buyerIndex].price = transactionPrice
		s[sellerIndex].price = transactionPrice

		// execute trade
		b[buyerIndex].quantityHeld = 1
		s[sellerIndex].quantityHeld = 0
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
	// seed RNG
	rand.Seed(time.Now().UTC().UnixNano())

	buyers, sellers := initializeAgents()
	openMarket(buyers, sellers)
}
