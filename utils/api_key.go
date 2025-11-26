package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func GenerateAPIKey(tenantID uint) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	randomPart := hex.EncodeToString(bytes)
	apiKey := fmt.Sprintf("%d_%s", tenantID, randomPart)
	return apiKey, nil
}

func ExtractTenantIDFromAPIKey(apiKey string) (uint, error) {
	parts := strings.Split(apiKey, "_")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid API key format")
	}
	var tenantID uint
	_, err := fmt.Sscanf(parts[0], "%d", &tenantID)
	if err != nil {
		return 0, fmt.Errorf("invalid tenant ID in API key")
	}
	return tenantID, nil
}

func ValidateAPIKeyFormat(apiKey string) bool {
	parts := strings.Split(apiKey, "_")
	if len(parts) != 2 {
		return false
	}
	var tenantID uint
	_, err := fmt.Sscanf(parts[0], "%d", &tenantID)
	if err != nil {
		return false
	}
	if len(parts[1]) != 64 {
		return false
	}
	_, err = hex.DecodeString(parts[1])
	return err == nil
}

func GetAPIKeyExpiry() time.Time {
	return time.Now().Add(10 * time.Minute)
}

func IsAPIKeyExpired(expiryTime *time.Time) bool {
	if expiryTime == nil {
		return true
	}
	return time.Now().After(*expiryTime)
}
