package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/api/idtoken"

	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
)

const IAPHeader = "x-goog-iap-jwt-assertion"

var ErrIAPAuth = errors.New("iap auth error")

type IAPUser struct {
	Email string
	Sub   string
}

type IAPValidator struct {
	settings *config.Settings
}

func NewIAPValidator(settings *config.Settings) *IAPValidator {
	return &IAPValidator{settings: settings}
}

func (v *IAPValidator) devBypassUser() *IAPUser {
	if v.settings.IAPJWTValidationDisabled && v.settings.Environment != "prod" {
		return &IAPUser{Email: "local-dev@internal", Sub: "local-dev"}
	}
	return nil
}

func (v *IAPValidator) Validate(ctx context.Context, token string) (*IAPUser, error) {
	if user := v.devBypassUser(); user != nil {
		return user, nil
	}

	if v.settings.IAPAudience == "" {
		return nil, fmt.Errorf("%w: IAP audience not configured", ErrIAPAuth)
	}

	payload, err := idtoken.Validate(ctx, token, v.settings.IAPAudience)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid IAP JWT", ErrIAPAuth)
	}

	email, _ := payload.Claims["email"].(string)
	sub, _ := payload.Claims["sub"].(string)
	if email == "" || sub == "" {
		return nil, fmt.Errorf("%w: IAP JWT missing identity claims", ErrIAPAuth)
	}

	verified, _ := payload.Claims["email_verified"].(bool)
	if !verified {
		return nil, fmt.Errorf("%w: IAP JWT email is not verified", ErrIAPAuth)
	}

	allowed := v.settings.ParsedAllowedEmailDomains()
	if len(allowed) > 0 {
		domain := emailDomain(email)
		hostedDomain, _ := payload.Claims["hd"].(string)
		if !contains(allowed, domain) && !contains(allowed, strings.ToLower(hostedDomain)) {
			return nil, fmt.Errorf("%w: email domain is not allowed", ErrIAPAuth)
		}
	}

	return &IAPUser{Email: strings.ToLower(email), Sub: sub}, nil
}

func emailDomain(email string) string {
	_, domain, ok := strings.Cut(email, "@")
	if !ok {
		return ""
	}
	return strings.ToLower(domain)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
