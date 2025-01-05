package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Grid dimensions
const (
	Width            = 50
	Height           = 50
	SellerSize       = 10
	TransportCost    = 1
	PriceSensitivity = 0.3
	NumberOfSellers  = 10
)

// Seller struct
type Seller struct {
	X, Y                   int
	Revenue                int
	MovementAggressiveness float64
	Color                  color.Color
	Price                  int
}

// Game struct
type Game struct {
	sellers    []Seller
	lastUpdate time.Time
}

// Random color
func randomColor() color.Color {
	return color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 0xff}
}

// Initialize sellers
func initializeSellers(numSellers int) []Seller {
	sellers := make([]Seller, numSellers)
	for i := range sellers {
		sellers[i] = Seller{
			X:                      rand.Intn(Width),
			Y:                      rand.Intn(Height),
			Revenue:                0,
			MovementAggressiveness: 1,
			Color:                  randomColor(),
			Price:                  20,
		}
	}
	return sellers
}

// Calculate distance between two points using Euclidean distance
func distance(x1, y1 int, seller Seller) float64 {
	x2, y2 := seller.X, seller.Y
	transportCost := TransportCost * math.Sqrt(math.Pow(float64(x1-x2), 2)+math.Pow(float64(y1-y2), 2))
	priceCost := PriceSensitivity * math.Abs(float64(seller.Price))
	return transportCost + priceCost
}

// Find closest seller
func findClosestSeller(sellers []Seller, x, y int) *Seller {
	var closest *Seller
	minDistance := math.MaxFloat64
	for i := range sellers {
		d := distance(x, y, sellers[i])
		if d < minDistance {
			minDistance = d
			closest = &sellers[i]
		}
	}
	return closest
}

// Simulate a day of sales
func simulateDay(sellers []Seller) {
	for x := 0; x < Width; x++ {
		for y := 0; y < Height; y++ {
			closest := findClosestSeller(sellers, x, y)
			closest.Revenue += closest.Price
		}
	}
}

// Calculate gradient and move seller towards higher paid region
func moveSellers(sellers []Seller) {
	for i := range sellers {
		if sellers[i].MovementAggressiveness < 0.05 {
			sellers[i].MovementAggressiveness = 0
		}

		if rand.Float64() > sellers[i].MovementAggressiveness {
			continue
		}

		// sellers[i].MovementAggressiveness *= 0.99

		originalX, originalY, originalPrice := sellers[i].X, sellers[i].Y, sellers[i].Price
		bestX, bestY, bestPrice := sellers[i].X, sellers[i].Y, sellers[i].Price
		bestRevenue := -1

		occupiedSpaces := make(map[[2]int]bool)
		for j := range sellers {
			if i == j {
				continue
			}
			occupiedSpaces[[2]int{sellers[j].X, sellers[j].Y}] = true
		}

		// lets say you can only move up, down, left, right
		moves := [][3]int{{0, 1, -1}, {0, 1, 0}, {0, 1, 1},
			{1, 0, -1}, {1, 0, 0}, {1, 0, 1},
			{-1, 0, -1}, {-1, 0, 0}, {-1, 0, 1},
			{0, -1, -1}, {0, -1, 0}, {0, -1, 1}}
		for _, move := range moves {
			dx, dy, dPrice := move[0], move[1], move[2]
			// cannot enter a space if there is already a seller there
			if occupiedSpaces[[2]int{originalX + dx, originalY + dy}] {
				continue
			}

			// cannot have a negative price
			if originalPrice+dPrice < 0 {
				continue
			}

			sellers[i].X, sellers[i].Y, sellers[i].Price = originalX+dx, originalY+dy, originalPrice+dPrice
			if sellers[i].X >= 0 && sellers[i].X < Width && sellers[i].Y >= 0 && sellers[i].Y < Height {
				sellers[i].Revenue = 0
				simulateDay(sellers)
				if sellers[i].Revenue > bestRevenue {
					bestX, bestY, bestPrice = sellers[i].X, sellers[i].Y, sellers[i].Price
					bestRevenue = sellers[i].Revenue
				}
			}
		}

		sellers[i].X, sellers[i].Y, sellers[i].Price = bestX, bestY, bestPrice
	}
}

// Update game state
func (g *Game) Update() error {
	moveSellers(g.sellers)
	g.lastUpdate = time.Now()
	return nil
}

// Draw game state
func (g *Game) Draw(screen *ebiten.Image) {
	for x := 0; x < Width; x++ {
		for y := 0; y < Height; y++ {
			seller := findClosestSeller(g.sellers, x, y)
			rgba := seller.Color.(color.RGBA)

			rgba.R /= 2
			rgba.G /= 2
			rgba.B /= 2

			ebitenutil.DrawRect(screen, float64(x)*SellerSize, float64(y)*SellerSize, SellerSize, SellerSize, rgba)
		}
	}

	// get the min price of sellers
	minPrice, maxPrice := math.MaxInt64, 0
	for _, seller := range g.sellers {
		if seller.Price < minPrice {
			minPrice = seller.Price
		}
		if seller.Price > maxPrice {
			maxPrice = seller.Price
		}
	}

	for _, seller := range g.sellers {

		ebitenutil.DrawRect(screen, float64(seller.X*SellerSize), float64(seller.Y*SellerSize), SellerSize, SellerSize, seller.Color)
		// draw the price inside the seller
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", seller.Price), seller.X*SellerSize, seller.Y*SellerSize)
	}
}

// Layout specifies the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return Width * SellerSize, Height * SellerSize
}

func main() {

	rand.Seed(time.Now().UnixNano())
	sellers := initializeSellers(NumberOfSellers)
	game := &Game{sellers: sellers}

	ebiten.SetWindowSize(Width*SellerSize, Height*SellerSize)
	ebiten.SetWindowTitle("2D Seller Simulation")

	if err := ebiten.RunGame(game); err != nil {
		fmt.Println(err)
	}
}
