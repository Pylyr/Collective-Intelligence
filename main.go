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
	Width      = 50
	Height     = 50
	SellerSize = 5
)

// Seller struct
type Seller struct {
	X, Y                   float64
	Points                 int
	MovementAggressiveness float64
	Color                  color.Color
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
			X:                      float64(rand.Intn(Width)),
			Y:                      float64(rand.Intn(Height)),
			Points:                 0,
			MovementAggressiveness: rand.Float64(),
			Color:                  randomColor(),
		}
	}
	return sellers
}

// Calculate distance between two points
func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

// Find closest seller
func findClosestSeller(sellers []Seller, x, y float64) *Seller {
	var closest *Seller
	minDistance := math.MaxFloat64
	for i := range sellers {
		d := distance(x, y, sellers[i].X, sellers[i].Y)
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
			closest := findClosestSeller(sellers, float64(x), float64(y))
			closest.Points++
		}
	}
}

func isSpaceOccupied(sellers []Seller, x, y float64, excludeIndex int) bool {
	for i := range sellers {
		if i == excludeIndex {
			continue
		}
		if sellers[i].X == x && sellers[i].Y == y {
			return true
		}
	}
	return false
}

// Calculate gradient and move seller towards higher paid region
func moveSellers(sellers []Seller) {
	for i := range sellers {
		if rand.Float64() < sellers[i].MovementAggressiveness {
			continue
		}

		originalX, originalY := sellers[i].X, sellers[i].Y

		bestX, bestY := sellers[i].X, sellers[i].Y
		bestPoints := -1

		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				// cannot enter a space if there is already a seller there
				if isSpaceOccupied(sellers, originalX+float64(dx), originalY+float64(dy), i) {
					continue
				}

				sellers[i].X, sellers[i].Y = originalX+float64(dx), originalY+float64(dy)
				if sellers[i].X >= 0 && sellers[i].X < Width && sellers[i].Y >= 0 && sellers[i].Y < Height {
					sellers[i].Points = 0
					simulateDay(sellers)
					if sellers[i].Points > bestPoints {
						bestX, bestY = sellers[i].X, sellers[i].Y
						bestPoints = sellers[i].Points
					}
				}
			}
		}

		sellers[i].X, sellers[i].Y = bestX, bestY
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
	for _, seller := range g.sellers {

		ebitenutil.DrawRect(screen, seller.X*SellerSize, seller.Y*SellerSize, SellerSize, SellerSize, seller.Color)
	}
}

// Layout specifies the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return Width * SellerSize, Height * SellerSize
}

func main() {
	ebiten.SetMaxTPS(10) // 10 ticks per second

	rand.Seed(time.Now().UnixNano())
	numSellers := 20
	sellers := initializeSellers(numSellers)
	game := &Game{sellers: sellers}

	ebiten.SetWindowSize(Width*SellerSize, Height*SellerSize)
	ebiten.SetWindowTitle("2D Seller Simulation")

	if err := ebiten.RunGame(game); err != nil {
		fmt.Println(err)
	}
}
