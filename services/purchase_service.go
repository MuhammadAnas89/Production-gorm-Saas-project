package services

import (
	"errors"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type PurchaseService struct {
	// Field ki zaroorat nahi kyunki hum har method mein DB inject kar rahe hain
}

func NewPurchaseService() *PurchaseService {
	return &PurchaseService{}
}

// 1. Create Request
func (s *PurchaseService) CreateRequest(tenantDB *gorm.DB, tenantID uint, userID uint, req *models.PurchaseOrder) error {
	req.TenantID = tenantID
	req.RequestedBy = userID
	req.Status = models.POPending

	repo := repositories.NewPurchaseRepository(tenantDB)
	return repo.Create(req)
}

// 2. Update Request (Mistake 1 Fixed: Added tenantID)
func (s *PurchaseService) UpdateRequest(tenantDB *gorm.DB, tenantID uint, orderID uint, quantity int, price float64) error {
	repo := repositories.NewPurchaseRepository(tenantDB)

	// ✅ Fixed: Passing tenantID to prevent accessing other tenant's order
	order, err := repo.GetByID(orderID, tenantID)
	if err != nil {
		return err
	}

	if order.Status != models.PORejected {
		return errors.New("only rejected orders can be updated")
	}

	order.Quantity = quantity
	order.BuyPrice = price
	order.Status = models.POPending
	return repo.Update(order)
}

// 3. Purchaser Action (Mistake 2 Fixed: Added tenantID)
func (s *PurchaseService) PurchaserAction(tenantDB *gorm.DB, tenantID uint, orderID uint, purchaserID uint, action string) error {
	repo := repositories.NewPurchaseRepository(tenantDB)

	// ✅ Fixed: Passing tenantID
	order, err := repo.GetByID(orderID, tenantID)
	if err != nil {
		return err
	}

	if order.Status != models.POPending {
		return errors.New("order is not in pending state")
	}

	switch action {
	case "approve":
		order.Status = models.PODispatched
		order.ApprovedBy = &purchaserID
	case "reject":
		order.Status = models.PORejected
		order.ApprovedBy = &purchaserID
	default:
		return errors.New("invalid action")
	}

	return repo.Update(order)
}

// 4. Receive Order (Mistakes 3 & 4 Fixed)
func (s *PurchaseService) ReceiveOrder(tenantDB *gorm.DB, tenantID uint, orderID uint) error {

	return tenantDB.Transaction(func(tx *gorm.DB) error {
		txRepo := repositories.NewPurchaseRepository(tx)
		order, err := txRepo.GetByID(orderID, tenantID)
		if err != nil {
			return err
		}

		if order.Status != models.PODispatched {
			return errors.New("order must be dispatched before receiving")
		}

		order.Status = models.POReceived
		if err := txRepo.Update(order); err != nil {
			return err
		}

		invRepo := repositories.NewInventoryRepository(tx)
		return invRepo.UpdateStock(order.ProductID, tenantID, order.Quantity)
	})
}

func (s *PurchaseService) ListOrders(tenantDB *gorm.DB, tenantID uint, status string) ([]models.PurchaseOrder, error) {

	return repositories.NewPurchaseRepository(tenantDB).List(tenantID, status)
}
