package separate

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	"github.com/NpoolPlatform/cloud-hashing-apis/pkg/middleware/referral"
	commissionsetting "github.com/NpoolPlatform/cloud-hashing-apis/pkg/middleware/referral/setting"
	orderconst "github.com/NpoolPlatform/cloud-hashing-order/pkg/const"
	npool "github.com/NpoolPlatform/message/npool/cloud-hashing-apis"
	inspirepb "github.com/NpoolPlatform/message/npool/cloud-hashing-inspire"

	"golang.org/x/xerrors"
)

func getUserGoodCommissions(ctx context.Context, appID, userID string) ([]*npool.GoodCommission, error) {
	orders, err := referral.GetOrders(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get orders: %v", err)
	}

	settings, err := commissionsetting.GetAmountSettingsByAppUser(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get amount settings: %v", err)
	}

	commissions := []*npool.GoodCommission{}

	for _, order := range orders {
		if order.Order.Payment == nil || order.Order.Payment.State != orderconst.PaymentStateDone {
			continue
		}

		if order.Order.Order.CreateAt >= dayBeginning() {
			continue
		}

		setting := commissionsetting.GetOrderAmountSetting(settings, order)
		if setting == nil {
			continue
		}

		orderAmount := order.Order.Payment.Amount * order.Order.Payment.CoinUSDCurrency

		var commission *npool.GoodCommission
		for _, comm := range commissions {
			if comm.GoodID == order.Good.Good.Good.ID {
				commission = comm
				break
			}
		}
		if commission == nil {
			commission = &npool.GoodCommission{
				GoodID:     order.Good.Good.Good.ID,
				CoinTypeID: order.Good.Main.ID,
				CoinName:   order.Good.Main.Unit,
			}
			commissions = append(commissions, commission)
		}

		commission.Amount += orderAmount * float64(setting.Percent) / 100.0

		logger.Sugar().Infof("order %v amount %v percent %v user %v", order.Order.Order.ID, orderAmount, setting.Percent, userID)
	}

	return commissions, nil
}

func getOrderParentGoodCommissions(ctx context.Context, appID, userID string, roots, nexts []*inspirepb.AppPurchaseAmountSetting) ([]*npool.GoodCommission, error) {
	orders, err := referral.GetOrders(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get orders: %v", err)
	}

	commissions := []*npool.GoodCommission{}

	for _, order := range orders {
		amount := getOrderParentRebate(ctx, order, roots, nexts)
		if amount <= 0 {
			continue
		}

		var commission *npool.GoodCommission
		for _, comm := range commissions {
			if comm.GoodID == order.Good.Good.Good.ID {
				commission = comm
				break
			}
		}
		if commission == nil {
			commission = &npool.GoodCommission{
				GoodID:     order.Good.Good.Good.ID,
				CoinTypeID: order.Good.Main.ID,
				CoinName:   order.Good.Main.Unit,
			}
			commissions = append(commissions, commission)
		}

		commission.Amount += amount
	}

	return commissions, nil
}

func getDirectInviteeGoodCommissions(ctx context.Context, appID, userID string) ([]*npool.GoodCommission, error) {
	roots, err := commissionsetting.GetAmountSettingsByAppUser(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get amount settings: %v", err)
	}

	invitees, err := referral.GetLayeredInvitees(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get layered invitees: %v", err)
	}

	commissions := []*npool.GoodCommission{}

	for _, iv := range invitees {
		inviteeID, err := findRootInviter(ctx, userID, iv.InviteeID, invitees)
		if err != nil {
			return nil, xerrors.Errorf("fail find root inviter: %v", err)
		}

		nexts, err := commissionsetting.GetAmountSettingsByAppUser(ctx, iv.AppID, inviteeID)
		if err != nil {
			return nil, xerrors.Errorf("fail get amount settings: %v", err)
		}

		comms, err := getOrderParentGoodCommissions(ctx, iv.AppID, iv.InviteeID, roots, nexts)
		if err != nil {
			return nil, xerrors.Errorf("fail get good commissions: %v", err)
		}

		for _, comm := range comms {
			for _, commission := range commissions {
				if commission.GoodID == comm.GoodID {
					commission.Amount += comm.Amount
				}
			}
		}
	}

	return commissions, nil
}

func getSeparateGoodCommissions(ctx context.Context, appID, userID string) ([]*npool.GoodCommission, error) {
	commissions, err := getUserGoodCommissions(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get user good commissions: %v", err)
	}

	comms, err := getDirectInviteeGoodCommissions(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get invitees good commissions: %v", err)
	}

	for _, comm := range comms {
		found := false
		for _, commission := range commissions {
			if commission.GoodID == comm.GoodID {
				commission.Amount += comm.Amount
				found = true
			}
		}
		if !found {
			commissions = append(commissions, comm)
		}
	}

	for _, comm := range commissions {
		comm.AppID = appID
		comm.UserID = userID
	}

	return commissions, nil
}

func GetSeparateGoodCommissions(ctx context.Context, appID, userID string) ([]*npool.GoodCommission, error) {
	comms, err := getSeparateGoodCommissions(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get user good commissions: %v", err)
	}

	invitees, err := referral.GetInvitees(ctx, appID, userID)
	if err != nil {
		return nil, xerrors.Errorf("fail get invitees: %v", err)
	}

	for _, iv := range invitees {
		commissions, err := getSeparateGoodCommissions(ctx, appID, iv.InviteeID)
		if err != nil {
			return nil, xerrors.Errorf("fail get user good commissions: %v", err)
		}
		comms = append(comms, commissions...)
	}

	return comms, nil
}