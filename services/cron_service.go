package services

import (
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

var CronSchedular *cron.Cron

func StartCronJob() {
	CronSchedular = cron.New()

	_, err := CronSchedular.AddFunc("@every 1m", func() {
		log.Println("Cron Start ho gai ha , Expire check kr lo ")
		DisableOldTenant()
	})
	if err != nil {
		log.Println("Failed to schedule cron job:", err)
		return
	}

	CronSchedular.Start()
	log.Println("Cron Job Started Successfully")
}

func DisableOldTenant() {
	db := config.MasterDB
	var tenants []models.Tenant
	expireTime := time.Now().Add(-10 * time.Hour)
	err := db.Where("is_active = ? AND created_at <= ?", true, expireTime).Find(&tenants).Error
	if err != nil {
		log.Println("Error fetching old inactive tenants:", err)
		return
	}
	if len(tenants) == 0 {
		log.Println("Koi inactive tenant nahi mila jo disable karna ho")
		return
	}

	for _, tenant := range tenants {
		tenant.IsActive = false
		err := db.Save(&tenant).Error
		if err != nil {
			log.Printf("Error disabling tenant ID %d: %v", tenant.ID, err)
		} else {
			cacheKey := fmt.Sprintf("tenant_info:%d", tenant.ID)
			_ = config.DeleteCache(cacheKey)

			log.Printf("Tenant ID %d disabled successfully", tenant.ID)
		}
	}
}
