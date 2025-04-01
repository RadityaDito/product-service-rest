package utils

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	adjectives = []string{
		"Awesome", "Cool", "Smart", "Innovative", "Premium",
		"Classic", "Elegant", "Advanced", "Ultimate", "Pro",
	}

	productTypes = []string{
		"Gadget", "Device", "Tool", "Accessory", "Electronics",
		"Appliance", "Instrument", "Machine", "Equipment", "System",
	}
)

type RandomProduct struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       float64
}

func GenerateRandomProduct() RandomProduct {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	adj := adjectives[rand.Intn(len(adjectives))]
	prodType := productTypes[rand.Intn(len(productTypes))]

	name := adj + " " + prodType
	description := "A " + strings.ToLower(name) + " designed for modern needs."
	price := 10.0 + rand.Float64()*990.0 // Price between 10 and 1000

	return RandomProduct{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Price:       price,
	}
}

func GenerateRandomProducts(count int) []RandomProduct {
	products := make([]RandomProduct, count)
	for i := 0; i < count; i++ {
		products[i] = GenerateRandomProduct()
	}
	return products
}
