package product_repository

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	product "github.com/pbb/orders-service/internal/product/model"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}
func (r *ProductRepository) CheckStock(productID int, quantity int) (bool, error) {
	var quantityScan int
	err := r.db.Table("products").Where("id = ?", productID).Select("quantity").Scan(&quantityScan).Error
	if err != nil {
		return false, err
	}
	return quantityScan >= quantity, nil
}

func (r *ProductRepository) DecreaseStock(productID int, quantity int) error {
	return r.db.Model(&product.Product{}).Where("id = ?", productID).Update("quantity", gorm.Expr("quantity - ?", quantity)).Error
}

func (r *ProductRepository) CheckAndUpdateProductQuantityWithTrans(productIDs []int, quantities []int) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction when processing products: %v", productIDs)
		return tx.Error
	}

	for i, productID := range productIDs {
		var quantity int

		err := tx.Table("products").Where("id = ?", productID).Select("quantity").Scan(&quantity).Error
		if err != nil {
			log.Printf("Failed to check stock for product %d: %v", productID, err)
			tx.Rollback() // Rollback the transaction on error
			return err
		}

		if quantity < quantities[i] {
			log.Printf("Insufficient stock for product %d. Available: %d, Requested: %d", productID, quantity, quantities[i])
			tx.Rollback()
			return fmt.Errorf("insufficient stock for product %d", productID)
		}

		err = tx.Model(&product.Product{}).Where("id = ?", productID).
			UpdateColumn("quantity", gorm.Expr("quantity - ?", quantities[i])).Error
		if err != nil {
			log.Printf("Failed to update stock for product %d: %v", productID, err)
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction for products: %v", productIDs)
		return err
	}

	log.Printf("Successfully checked and updated stock for products: %v", productIDs)
	return nil
}

func (r *ProductRepository) UpdateInventory(productID int, quantity int) error {
	if err := r.db.Model(&product.Product{}).
		Where("id = ?", productID).
		UpdateColumn("quantity", gorm.Expr("quantity - ?", quantity)).Error; err != nil {
		log.Printf("Failed to update inventory for product %d: %v", productID, err)
		return err
	}

	log.Printf("Successfully updated inventory for product %d", productID)
	return nil
}
