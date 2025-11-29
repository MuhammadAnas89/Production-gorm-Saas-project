package services

import (
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gopkg.in/gomail.v2"
)

type CronService struct {
	cron *cron.Cron
}

func NewCronService() *CronService {
	c := cron.New()
	return &CronService{cron: c}
}

func (s *CronService) Start() {
	// 1. Plan Expiry Check (Rozana Raat 12 Bajay)
	_, err := s.cron.AddFunc("0 0 * * *", func() {
		log.Println("⏰ CRON: Checking Plan Expiries...")
		s.HandlePlanExpiries()
	})
	if err != nil {
		log.Printf("Error scheduling expiry job: %v", err)
	}

	// 2. Low Stock Alert (Rozana Subah 9 Bajay)
	_, err = s.cron.AddFunc("0 9 * * *", func() {
		log.Println("⏰ CRON: Checking Low Stock...")
		s.CheckLowStock()
	})
	if err != nil {
		log.Printf("Error scheduling stock job: %v", err)
	}

	s.cron.Start()
	log.Println("✅ Cron Scheduler Started")
}

func (s *CronService) Stop() {
	s.cron.Stop()
}

// ---------------------------------------------------------
// JOB 1: Handle Plan Expiry (Suspend + Warning Email)
// ---------------------------------------------------------
func (s *CronService) HandlePlanExpiries() {
	db := config.GetMasterDB()
	var tenants []models.Tenant

	now := time.Now()
	warningThreshold := now.Add(3 * 24 * time.Hour) // 3 Days later

	// Filter: Sirf wo tenants jo ACTIVE hain aur jinki expiry set hai
	if err := db.Preload("Plan").Where("is_active = ? AND plan_expiry IS NOT NULL", true).Find(&tenants).Error; err != nil {
		log.Printf("Error fetching tenants: %v", err)
		return
	}

	for _, tenant := range tenants {

		// A. Check if EXPIRED (Suspension Logic)
		if tenant.PlanExpiry.Before(now) {
			log.Printf("🚫 Tenant %s expired. Suspending account...", tenant.Name)

			// ✅ CHANGE: Inactive kar do (Downgrade nahi)
			tenant.IsActive = false
			db.Save(&tenant)

			// 3. Cache Clear karo (Taaki wo foran logout ho jaye)
			cacheService := NewCacheService()
			cacheService.Delete(fmt.Sprintf("tenant_info:%d", tenant.ID))

			// Notify Admin
			email, _ := s.GetTenantAdminEmail(&tenant)
			if email != "" {
				s.SendEmail(email, "Account Suspended - Plan Expired",
					fmt.Sprintf("Your subscription for <b>%s</b> has expired. Your access has been suspended. Please renew your plan to continue.", tenant.Name))
			}
			continue
		}

		// B. Check if EXPIRING SOON (Warning Logic)
		if tenant.PlanExpiry.Before(warningThreshold) {
			log.Printf("⚠️ Tenant %s expiring soon. Sending warning...", tenant.Name)

			email, _ := s.GetTenantAdminEmail(&tenant)
			if email != "" {
				daysLeft := int(time.Until(*tenant.PlanExpiry).Hours() / 24)
				s.SendEmail(email, "Action Required: Plan Expiring Soon",
					fmt.Sprintf("Your plan for <b>%s</b> will expire in %d days. If not renewed, your account will be suspended.", tenant.Name, daysLeft))
			}
		}
	}
}

// ---------------------------------------------------------
// JOB 2: Low Stock Alerts
// ---------------------------------------------------------
func (s *CronService) CheckLowStock() {
	// Sirf Active tenants ka stock check karo
	var tenants []models.Tenant
	if err := config.GetMasterDB().Where("is_active = ?", true).Find(&tenants).Error; err != nil {
		return
	}

	for _, tenant := range tenants {
		tenantDB, err := config.TenantManager.GetTenantDB(&tenant)
		if err != nil {
			continue
		}

		repo := repositories.NewInventoryRepository(tenantDB)
		items, err := repo.GetLowStockProducts(10)

		if err == nil && len(items) > 0 {
			email, _ := s.GetTenantAdminEmail(&tenant)
			if email != "" {
				s.SendEmail(email, "Low Stock Alert",
					fmt.Sprintf("You have %d items running low on stock. Please check your inventory.", len(items)))
			}
		}
	}
}

// ---------------------------------------------------------
// Helpers
// ---------------------------------------------------------

func (s *CronService) GetTenantAdminEmail(tenant *models.Tenant) (string, error) {
	tenantDB, err := config.TenantManager.GetTenantDB(tenant)
	if err != nil {
		return "", err
	}

	var user models.User
	err = tenantDB.Table("users").
		Select("users.email").
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", "Tenant Admin").
		Limit(1).
		First(&user).Error

	return user.Email, err
}

func (s *CronService) SendEmail(to string, subject string, body string) {
	from := "muhammad.anas.khalid.13@gmail.com"
	pass := "atuyfmtqmsqmepxh"
	host := "smtp.gmail.com"
	port := 587

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", fmt.Sprintf("<div style='font-family: sans-serif; padding: 20px; border: 1px solid #ccc;'>%s</div>", body))

	d := gomail.NewDialer(host, port, from, pass)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
	} else {
		log.Printf("✅ Email sent to %s", to)
	}
}
