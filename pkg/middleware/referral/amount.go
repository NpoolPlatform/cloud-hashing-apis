package referral

import (
	"context"
	"fmt"

	cache "github.com/NpoolPlatform/cloud-hashing-apis/pkg/middleware/cache"
	orderconst "github.com/NpoolPlatform/cloud-hashing-order/pkg/const"

	"golang.org/x/xerrors"
)

const (
	cacheUSDAmount       = "referral:usd:amount"
	cachePeriodUSDAmount = "referral:period:usd:amount"
)

func GetUSDAmount(ctx context.Context, appID, userID string) (float64, error) {
	amount := cache.GetEntry(CacheKey(appID, userID, cacheUSDAmount))
	if amount != nil {
		return *(amount.(*float64)), nil
	}

	// TODO: let database to sum orders amount
	orders, err := GetOrders(ctx, appID, userID)
	if err != nil {
		return 0, xerrors.Errorf("fail get orders: %v", err)
	}

	totalAmount := 0.0
	for _, order := range orders {
		if order.Payment == nil || order.Payment.State != orderconst.PaymentStateDone {
			continue
		}
		totalAmount += order.Payment.Amount * order.Payment.CoinUSDCurrency
	}

	cache.AddEntry(CacheKey(appID, userID, cacheUSDAmount), &totalAmount)

	return totalAmount, nil
}

func GetSubUSDAmount(ctx context.Context, appID, userID string) (float64, error) {
	invitees, err := GetLayeredInvitees(ctx, appID, userID)
	if err != nil {
		return 0, xerrors.Errorf("fail get invitees: %v", err)
	}

	totalAmount := 0.0
	for _, iv := range invitees {
		amount, err := GetUSDAmount(ctx, iv.AppID, iv.InviteeID)
		if err != nil {
			return 0, xerrors.Errorf("fail get usd amount: %v", err)
		}
		totalAmount += amount
	}

	return totalAmount, nil
}

func GetPeriodUSDAmount(ctx context.Context, appID, userID string, start, end uint32) (float64, error) {
	key := fmt.Sprintf("%v:%v:%v", CacheKey(appID, userID, cachePeriodUSDAmount), start, end)
	amount := cache.GetEntry(key)
	if amount != nil {
		return *(amount.(*float64)), nil
	}

	orders, err := GetOrders(ctx, appID, userID)
	if err != nil {
		return 0, xerrors.Errorf("fail get orders: %v", err)
	}

	totalAmount := 0.0
	for _, order := range orders {
		if order.Payment == nil || order.Payment.State != orderconst.PaymentStateDone {
			continue
		}

		if order.Order.CreateAt < start || (order.Order.CreateAt >= end && end > 0) {
			continue
		}
		totalAmount += order.Payment.Amount * order.Payment.CoinUSDCurrency
	}

	cache.AddEntry(key, &totalAmount)

	return totalAmount, nil
}

func GetPeriodSubUSDAmount(ctx context.Context, appID, userID string, start, end uint32) (float64, error) {
	invitees, err := GetLayeredInvitees(ctx, appID, userID)
	if err != nil {
		return 0, xerrors.Errorf("fail get invitees: %v", err)
	}

	totalAmount := 0.0
	for _, iv := range invitees {
		amount, err := GetPeriodUSDAmount(ctx, iv.AppID, iv.InviteeID, start, end)
		if err != nil {
			return 0, xerrors.Errorf("fail get usd amount: %v", err)
		}
		totalAmount += amount
	}

	return totalAmount, nil
}
