package seed

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"muma-ramen-backend/models"
)

func SeedDatabase(db *gorm.DB) {
	// Check if data already exists
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database...")

	// Create Super Admin
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("superadmin123"), bcrypt.DefaultCost)
	superAdmin := models.User{
		Name:         "Super Admin",
		Email:        "superadmin@mumaramen.com",
		PasswordHash: string(hashedPassword),
		Role:         models.RoleSuperAdmin,
	}
	db.Create(&superAdmin)

	// Create Admin
	hashedPassword, _ = bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := models.User{
		Name:         "Admin Muma",
		Email:        "admin@mumaramen.com",
		PasswordHash: string(hashedPassword),
		Role:         models.RoleAdmin,
	}
	db.Create(&admin)

	// Create Kasir
	hashedPassword, _ = bcrypt.GenerateFromPassword([]byte("kasir123"), bcrypt.DefaultCost)
	kasir := models.User{
		Name:         "Kasir 1",
		Email:        "kasir@mumaramen.com",
		PasswordHash: string(hashedPassword),
		Role:         models.RoleKasir,
	}
	db.Create(&kasir)

	// Create sample Pembeli
	hashedPassword, _ = bcrypt.GenerateFromPassword([]byte("pembeli123"), bcrypt.DefaultCost)
	pembeli := models.User{
		Name:         "Pembeli Demo",
		Email:        "pembeli@mumaramen.com",
		PasswordHash: string(hashedPassword),
		Role:         models.RolePembeli,
	}
	db.Create(&pembeli)

	// Create Categories
	categories := []models.Category{
		{Name: "Ramen", Description: "Signature ramen bowls", SortOrder: 1},
		{Name: "Side Dishes", Description: "Pelengkap ramen favorit", SortOrder: 2},
		{Name: "Rice Bowl", Description: "Nasi dengan topping spesial", SortOrder: 3},
		{Name: "Appetizers", Description: "Makanan pembuka", SortOrder: 4},
		{Name: "Drinks", Description: "Minuman segar", SortOrder: 5},
		{Name: "Desserts", Description: "Penutup manis", SortOrder: 6},
	}
	for i := range categories {
		db.Create(&categories[i])
	}

	// Create Menu Items
	menuItems := []models.MenuItem{
		// Ramen (Category 1)
		{CategoryID: categories[0].ID, Name: "Tonkotsu Ramen", Description: "Kuah babi kental yang gurih dengan chashu, jamur kikurage, nori, dan telur ajitama", Price: 55000, ImageURL: "/images/tonkotsu.jpg", IsAvailable: true, Stock: 50},
		{CategoryID: categories[0].ID, Name: "Miso Ramen", Description: "Kuah miso yang kaya rasa dengan jagung manis, mentega, dan moyashi", Price: 50000, ImageURL: "/images/miso.jpg", IsAvailable: true, Stock: 50},
		{CategoryID: categories[0].ID, Name: "Shoyu Ramen", Description: "Kuah kecap Jepang yang bening dengan chashu, menma, dan daun bawang", Price: 48000, ImageURL: "/images/shoyu.jpg", IsAvailable: true, Stock: 50},
		{CategoryID: categories[0].ID, Name: "Spicy Tantanmen", Description: "Ramen pedas dengan daging cincang, bok choy, dan wijen", Price: 58000, ImageURL: "/images/tantanmen.jpg", IsAvailable: true, Stock: 40},
		{CategoryID: categories[0].ID, Name: "Curry Ramen", Description: "Perpaduan kuah kare Jepang dengan ramen, topping katsu ayam", Price: 60000, ImageURL: "/images/curry-ramen.jpg", IsAvailable: true, Stock: 35},
		{CategoryID: categories[0].ID, Name: "Black Garlic Ramen", Description: "Tonkotsu dengan minyak bawang hitam, telur onsen, dan chashu panggang", Price: 62000, ImageURL: "/images/black-garlic.jpg", IsAvailable: true, Stock: 30},

		// Side Dishes (Category 2)
		{CategoryID: categories[1].ID, Name: "Gyoza (5 pcs)", Description: "Pangsit goreng isi daging ayam dan sayuran", Price: 28000, ImageURL: "/images/gyoza.jpg", IsAvailable: true, Stock: 80},
		{CategoryID: categories[1].ID, Name: "Chicken Karaage", Description: "Ayam goreng tepung khas Jepang dengan mayo", Price: 32000, ImageURL: "/images/karaage.jpg", IsAvailable: true, Stock: 60},
		{CategoryID: categories[1].ID, Name: "Takoyaki (6 pcs)", Description: "Bola-bola gurita dengan saus takoyaki dan bonito flakes", Price: 25000, ImageURL: "/images/takoyaki.jpg", IsAvailable: true, Stock: 45},
		{CategoryID: categories[1].ID, Name: "Edamame", Description: "Kacang kedelai rebus dengan garam laut", Price: 18000, ImageURL: "/images/edamame.jpg", IsAvailable: true, Stock: 100},

		// Rice Bowl (Category 3)
		{CategoryID: categories[2].ID, Name: "Chashu Don", Description: "Nasi dengan chashu babi panggang, telur onsen, dan saus teriyaki", Price: 45000, ImageURL: "/images/chashu-don.jpg", IsAvailable: true, Stock: 40},
		{CategoryID: categories[2].ID, Name: "Chicken Katsu Don", Description: "Nasi dengan ayam katsu, telur, dan saus donburi", Price: 42000, ImageURL: "/images/katsu-don.jpg", IsAvailable: true, Stock: 40},
		{CategoryID: categories[2].ID, Name: "Beef Gyudon", Description: "Nasi dengan irisan daging sapi dan bawang bombay manis", Price: 48000, ImageURL: "/images/gyudon.jpg", IsAvailable: true, Stock: 35},

		// Appetizers (Category 4)
		{CategoryID: categories[3].ID, Name: "Ebi Tempura (3 pcs)", Description: "Udang goreng tepung renyah dengan saus tentsuyu", Price: 35000, ImageURL: "/images/tempura.jpg", IsAvailable: true, Stock: 50},
		{CategoryID: categories[3].ID, Name: "Agedashi Tofu", Description: "Tahu goreng dengan kuah dashi hangat dan katsuobushi", Price: 22000, ImageURL: "/images/agedashi.jpg", IsAvailable: true, Stock: 40},

		// Drinks (Category 5)
		{CategoryID: categories[4].ID, Name: "Ocha (Green Tea)", Description: "Teh hijau Jepang panas/dingin", Price: 12000, ImageURL: "/images/ocha.jpg", IsAvailable: true, Stock: 200},
		{CategoryID: categories[4].ID, Name: "Calpico", Description: "Minuman yogurt Jepang segar", Price: 15000, ImageURL: "/images/calpico.jpg", IsAvailable: true, Stock: 100},
		{CategoryID: categories[4].ID, Name: "Yuzu Lemonade", Description: "Lemonade segar dengan perasan yuzu", Price: 20000, ImageURL: "/images/yuzu.jpg", IsAvailable: true, Stock: 80},
		{CategoryID: categories[4].ID, Name: "Ramune Soda", Description: "Soda Jepang klasik berbagai rasa", Price: 18000, ImageURL: "/images/ramune.jpg", IsAvailable: true, Stock: 60},

		// Desserts (Category 6)
		{CategoryID: categories[5].ID, Name: "Matcha Ice Cream", Description: "Es krim matcha premium dengan mochi", Price: 22000, ImageURL: "/images/matcha-ice.jpg", IsAvailable: true, Stock: 40},
		{CategoryID: categories[5].ID, Name: "Dorayaki", Description: "Pancake Jepang isi kacang merah", Price: 18000, ImageURL: "/images/dorayaki.jpg", IsAvailable: true, Stock: 50},
		{CategoryID: categories[5].ID, Name: "Mochi (3 pcs)", Description: "Mochi lembut isi matcha, strawberry, dan cokelat", Price: 20000, ImageURL: "/images/mochi.jpg", IsAvailable: true, Stock: 60},
	}

	for i := range menuItems {
		db.Create(&menuItems[i])
	}

	fmt.Println("✅ Database seeded successfully!")
	fmt.Println("📧 Login accounts:")
	fmt.Println("   Super Admin: superadmin@mumaramen.com / superadmin123")
	fmt.Println("   Admin:       admin@mumaramen.com / admin123")
	fmt.Println("   Kasir:       kasir@mumaramen.com / kasir123")
	fmt.Println("   Pembeli:     pembeli@mumaramen.com / pembeli123")
}
