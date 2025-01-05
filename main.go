package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	Width            = 50
	Height           = 50
	SellerSize       = 15
	TransportCost    = 1
	PriceSensitivity = 1
	MaxPrice         = 20
	StartPrice       = 10
	NSellers         = 2
)

// Graph constants
const (
	GraphHeight = Height*SellerSize - 100
	BarWidth    = 20
)

// Seller struct
type Seller struct {
	X, Y                   int
	Revenue                float64
	MovementAggressiveness float64
	Color                  color.Color
	Price                  int
}

// Game struct
type Game struct {
	sellers []Seller
	turns   int
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

func distance(x1, y1 int, seller Seller) float64 {
	x2, y2 := seller.X, seller.Y

	// Euclidean distance for transport cost
	transportCost := TransportCost * math.Sqrt(math.Pow(float64(x1-x2), 2)+math.Pow(float64(y1-y2), 2))

	// Non-linear price sensitivity using a logarithmic function
	priceCost := PriceSensitivity * float64(seller.Price)

	// Total cost = transport cost + price cost
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
			closest.Revenue += float64(closest.Price)
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
		bestRevenue := -1.0

		occupiedSpaces := make(map[[2]int]bool)
		for j := range sellers {
			if i == j {
				continue
			}
			occupiedSpaces[[2]int{sellers[j].X, sellers[j].Y}] = true
		}

		// lets say you can only move up, down, left, right
		// moves := [][3]int{{0, 1, -1}, {0, 1, 0}, {0, 1, 1},
		// 	{1, 0, -1}, {1, 0, 0}, {1, 0, 1},
		// 	{-1, 0, -1}, {-1, 0, 0}, {-1, 0, 1},
		// 	{0, -1, -1}, {0, -1, 0}, {0, -1, 1},
		// 	{0, 0, -1}, {0, 0, 0}, {0, 0, 1}}
		moves := [][3]int{{0, 1, 0}, {1, 0, 0}, {-1, 0, 0}, {0, -1, 0}, {0, 0, -1}, {0, 0, 0}, {0, 0, 1}}
		for _, move := range moves {
			dx, dy, dPrice := move[0], move[1], move[2]
			// cannot enter a space if there is already a seller there
			if occupiedSpaces[[2]int{originalX + dx, originalY + dy}] {
				continue
			}

			// cannot have a negative price
			if originalPrice+dPrice < 0 || originalPrice+dPrice > MaxPrice {
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
		sellers[i].Revenue = bestRevenue
	}
}

func (g *Game) saveScreenshot() {
	img := ebiten.NewImage(Width*SellerSize+200+len(g.sellers)*BarWidth, Height*SellerSize)
	g.Draw(img)

	// make a name that reflects the run
	fileName := fmt.Sprintf("Turn_%d_NS%d_TC_%d_PS_%d_MP_%d_SP_%d.png", g.turns, NSellers, TransportCost, PriceSensitivity, MaxPrice, StartPrice)

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		fmt.Println("Failed to encode image:", err)
		return
	}

	fmt.Println("Screenshot saved as simulation_turn_100.png")
}

// Update game state
func (g *Game) Update() error {
	moveSellers(g.sellers)
	g.turns++

	if g.turns%100 == 0 {
		g.saveScreenshot()
	}

	return nil
}

func drawText(screen *ebiten.Image, text string, x, y int) {
	face := basicfont.Face7x13

	d := &font.Drawer{
		Dst:  screen,
		Src:  image.White,
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

func (g *Game) DrawSimulation(screen *ebiten.Image) {
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
func (g *Game) DrawRevenueGraph(screen *ebiten.Image) {
	// Calculate everyone's revenue
	for i := range g.sellers {
		g.sellers[i].Revenue = 0
	}
	simulateDay(g.sellers)

	// Scale factor to fit the graph within the screen
	maxRevenue := 0.0
	for _, seller := range g.sellers {
		if seller.Revenue > maxRevenue {
			maxRevenue = seller.Revenue
		}
	}

	scaleFactor := float64(GraphHeight) / float64(maxRevenue)
	xOffset := Width*SellerSize + 100

	// Draw the title
	title := "Daily Revenue per Seller"
	// make the text big
	drawText(screen, title, xOffset-50, 30)

	// Draw bars
	for i, seller := range g.sellers {
		barHeight := float64(seller.Revenue) * scaleFactor
		barX := xOffset + i*BarWidth
		barY := GraphHeight - int(barHeight) + 60 // Adjust for title space
		rgba := seller.Color.(color.RGBA)
		rgba.R /= 2
		rgba.G /= 2
		rgba.B /= 2

		ebitenutil.DrawRect(screen, float64(barX), float64(barY), float64(BarWidth), barHeight, rgba)
	}

	// Draw scale
	for i := 0; i <= 5; i++ {
		scaleValue := int(float64(maxRevenue) * float64(i) / 5)
		y := GraphHeight - int(float64(i*GraphHeight/5)) + 60 // Adjust for title space

		// Draw tick lines
		ebitenutil.DrawLine(screen, float64(xOffset)-5, float64(y), float64(xOffset), float64(y), color.White)

		// Draw scale labels
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", scaleValue), xOffset-40, y-5)
	}
}

// Draw game state
func (g *Game) Draw(screen *ebiten.Image) {
	// make the background white
	// screen.Fill(color.Black)
	g.DrawSimulation(screen)
	g.DrawRevenueGraph(screen)
}

// Layout specifies the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return Width*SellerSize + 200 + len(g.sellers)*BarWidth, Height * SellerSize
}

func main() {

	rand.Seed(time.Now().UnixNano())
	sellers := initializeSellers(NSellers)
	game := &Game{sellers: sellers}

	ebiten.SetWindowSize(Width*SellerSize, Height*SellerSize)
	ebiten.SetWindowTitle("2D Seller Simulation")

	// Graph window
	if err := ebiten.RunGame(game); err != nil {
		fmt.Println(err)
	}

}
