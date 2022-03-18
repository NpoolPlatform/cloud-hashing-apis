package referral

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/cloud-hashing-apis"

	grpc2 "github.com/NpoolPlatform/cloud-hashing-apis/pkg/grpc"
	order "github.com/NpoolPlatform/cloud-hashing-apis/pkg/middleware/order"

	orderconst "github.com/NpoolPlatform/cloud-hashing-order/pkg/const"
	appusermgrpb "github.com/NpoolPlatform/message/npool/appusermgr"
	inspirepb "github.com/NpoolPlatform/message/npool/cloud-hashing-inspire"
	coininfopb "github.com/NpoolPlatform/message/npool/coininfo"

	"golang.org/x/xerrors"
)

var (
	appInvitations        = map[string]map[string]map[string]*npool.Invitation{}
	appInviterUserInfos   = map[string]map[string]*npool.InvitationUserInfo{}
	appInviterLastUpdates = map[string]map[string]time.Time{}
	notifier              = make(chan struct{})
	mutex                 = sync.Mutex{}
)

func AddWatcher(appID, inviterID string) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := appInvitations[appID]; !ok {
		appInvitations[appID] = map[string]map[string]*npool.Invitation{}
	}
	appInvitation := appInvitations[appID]
	if _, ok := appInvitation[inviterID]; !ok {
		appInvitation[inviterID] = map[string]*npool.Invitation{}
	}
	appInvitations[appID] = appInvitation

	if _, ok := appInviterUserInfos[appID]; !ok {
		appInviterUserInfos[appID] = map[string]*npool.InvitationUserInfo{}
	}
	appInviterUserInfo := appInviterUserInfos[appID]
	if _, ok := appInviterUserInfo[inviterID]; !ok {
		appInviterUserInfo[inviterID] = &npool.InvitationUserInfo{}
	}
	appInviterUserInfos[appID] = appInviterUserInfo

	if _, ok := appInviterLastUpdates[appID]; !ok {
		appInviterLastUpdates[appID] = map[string]time.Time{}
	}
	appInviterLastUpdate := appInviterLastUpdates[appID]
	if _, ok := appInviterLastUpdate[inviterID]; !ok {
		appInviterLastUpdate[inviterID] = time.Time{}
	}
	appInviterLastUpdates[appID] = appInviterLastUpdate

	go func() {
		notifier <- struct{}{}
	}()
}

func update() time.Duration {
	appInviters := map[string][]string{}

	mutex.Lock()
	for appID, inviterMap := range appInvitations {
		if _, ok := appInviters[appID]; !ok {
			appInviters[appID] = []string{}
		}
		myInviters := appInviters[appID]
		for inviterID := range inviterMap {
			myInviters = append(myInviters, inviterID)
		}
		appInviters[appID] = myInviters
	}
	mutex.Unlock()

	logger.Sugar().Infof("run async updater at %v", time.Now())
	toNext := 24 * time.Hour

	for appID, inviters := range appInviters {
		for _, inviterID := range inviters {
			nextSync := appInviterLastUpdates[appID][inviterID].Add(24 * time.Hour)
			if time.Now().Before(nextSync) {
				curToNext := time.Until(nextSync)
				if curToNext < toNext {
					toNext = curToNext
				}
				continue
			}

			logger.Sugar().Infof("run async updater for %v at %v", inviterID, time.Now())

			invitations, userInfo, err := getInvitations(appID, inviterID, false)
			if err != nil {
				logger.Sugar().Errorf("fail get invitations: %v", err)
				continue
			}

			mutex.Lock()
			appInvitations[appID][inviterID] = invitations
			appInviterUserInfos[appID][inviterID] = userInfo
			appInviterLastUpdates[appID][inviterID] = time.Now()
			mutex.Unlock()
		}
	}

	return toNext
}

func Run() {
	var timeout time.Duration

	timer := time.NewTimer(24 * time.Hour)

	for {
		select {
		case <-timer.C:
			timeout = update()
		case <-notifier:
			timeout = update()
		}
		timer.Reset(timeout)
	}
}

func GetCommission(appID, userID string) (float64, error) {
	mutex.Lock()
	invitations := appInvitations[appID][userID]
	userInfo := appInviterUserInfos[appID][userID]
	mutex.Unlock()

	if len(invitations) > 0 {
		return userInfo.CommissionAmount, nil
	}

	_, userInfo, err := getInvitations(appID, userID, false)
	if err != nil {
		return 0, xerrors.Errorf("fail get invitations: %v", err)
	}

	return userInfo.CommissionAmount, nil
}

func getFullInvitations(appID, inviterID string) (map[string]*npool.Invitation, *npool.InvitationUserInfo, error) {
	mutex.Lock()
	invitations := appInvitations[appID][inviterID]
	userInfo := appInviterUserInfos[appID][inviterID]
	mutex.Unlock()

	if len(invitations) > 0 {
		return invitations, userInfo, nil
	}

	invitations, userInfo, err := getInvitations(appID, inviterID, false)
	if err != nil {
		return nil, nil, xerrors.Errorf("fail get invitations: %v", err)
	}

	return invitations, userInfo, nil
}

func getDirectInvitations(appID, inviterID string) (map[string]*npool.Invitation, *npool.InvitationUserInfo, error) {
	mutex.Lock()
	invitations := appInvitations[appID][inviterID]
	userInfo := appInviterUserInfos[appID][inviterID]
	mutex.Unlock()

	if len(invitations) > 0 {
		return invitations, userInfo, nil
	}

	invitations, userInfo, err := getInvitations(appID, inviterID, true)
	if err != nil {
		return nil, nil, xerrors.Errorf("fail get invitations: %v", err)
	}

	return invitations, userInfo, nil
}

func getInvitationUserInfo( //nolint
	appID, inviteeID string,
	myGoods map[string]*npool.Good,
	myCoins map[string]*coininfopb.CoinInfo) (*npool.InvitationUserInfo,
	map[string]*npool.Good,
	map[string]*coininfopb.CoinInfo,
	error) {
	ctx := context.Background()

	inviteeResp, err := grpc2.GetAppUserInfoByAppUser(ctx, &appusermgrpb.GetAppUserInfoByAppUserRequest{
		AppID:  appID,
		UserID: inviteeID,
	})
	if err != nil {
		return nil, myGoods, myCoins, xerrors.Errorf("fail get invitee %v user info: %v", inviteeID, err)
	}

	resp1, err := grpc2.GetUserInvitationCodeByAppUser(ctx, &inspirepb.GetUserInvitationCodeByAppUserRequest{
		AppID:  appID,
		UserID: inviteeResp.Info.User.ID,
	})
	if err != nil {
		return nil, myGoods, myCoins, xerrors.Errorf("fail get user invitation code: %v", err)
	}

	summarys := map[string]*npool.InvitationSummary{}

	resp2, goods, coins, err := order.GetOrdersShortDetailByAppUser(ctx, &npool.GetOrdersByAppUserRequest{
		AppID:  appID,
		UserID: inviteeResp.Info.User.ID,
	}, myGoods, myCoins)
	if err != nil {
		return nil, myGoods, myCoins, xerrors.Errorf("fail get orders detail by app user: %v", err)
	}

	myGoods = goods
	myCoins = coins

	appCommissionSetting, err := grpc2.GetAppCommissionSettingByApp(ctx, &inspirepb.GetAppCommissionSettingByAppRequest{
		AppID: appID,
	})
	if err != nil {
		return nil, myGoods, myCoins, xerrors.Errorf("fail get app commission setting: %v", err)
	}

	commissions := []*npool.CommissionAmount{}

	if appCommissionSetting != nil {
		if appCommissionSetting.UniqueSetting {
			appPurchaseAmountSettings, err := grpc2.GetAppPurchaseAmountSettingsByApp(ctx, &inspirepb.GetAppPurchaseAmountSettingsByAppRequest{
				AppID: appID,
			})
			if err != nil {
				return nil, myGoods, myCoins, xerrors.Errorf("fail get app purchase amount setting: %v", err)
			}

			sort.Slice(appPurchaseAmountSettings.Infos, func(i, j int) bool {
				return appPurchaseAmountSettings.Infos[i].Amount < appPurchaseAmountSettings.Infos[j].Amount
			})

			for _, info := range appPurchaseAmountSettings.Infos {
				commissions = append(commissions, &npool.CommissionAmount{
					Amount:  info.Amount,
					Percent: info.Percent,
					Start:   info.Start,
					End:     info.End,
				})
			}
		} else {
			myInviteeSettings, err := grpc2.GetAppPurchaseAmountSettingsByAppUser(ctx, &inspirepb.GetAppPurchaseAmountSettingsByAppUserRequest{
				AppID:  appID,
				UserID: inviteeID,
			})
			if err != nil {
				return nil, myGoods, myCoins, xerrors.Errorf("fail get app purchase amount setting: %v", err)
			}

			sort.Slice(myInviteeSettings, func(i, j int) bool {
				return myInviteeSettings[i].Amount < myInviteeSettings[j].Amount
			})

			for _, info := range myInviteeSettings {
				commissions = append(commissions, &npool.CommissionAmount{
					Amount:  info.Amount,
					Percent: info.Percent,
					Start:   info.Start,
					End:     info.End,
				})
			}
		}
	}

	kol := false
	if resp1.Info != nil {
		kol = true
	}

	if !kol {
		commissions = []*npool.CommissionAmount{}
	}

	for orderIndex, orderInfo := range resp2.Infos {
		if orderInfo.Order.Payment == nil {
			continue
		}

		if orderInfo.Order.Payment.State != orderconst.PaymentStateDone {
			continue
		}

		if _, ok := summarys[orderInfo.Good.Good.Good.CoinInfoID]; !ok {
			summarys[orderInfo.Good.Good.Good.CoinInfoID] = &npool.InvitationSummary{}
		}

		summary := summarys[orderInfo.Good.Good.Good.CoinInfoID]
		summary.Units += orderInfo.Order.Order.Units

		usdAmount := orderInfo.Order.Payment.Amount
		if orderInfo.Order.Payment.CoinUSDCurrency > 0 {
			usdAmount *= orderInfo.Order.Payment.CoinUSDCurrency
		}

		if kol {
			for i, commission := range commissions {
				if commission.Start <= orderInfo.Order.Payment.CreateAt && orderInfo.Order.Payment.CreateAt < commission.End {
					commissions[i].PayAmount += usdAmount
					commissions[i].CreateAt = orderInfo.Order.Payment.CreateAt
					logger.Sugar().Infof("app %v user %v order %v | %v commission %v commission %v container %v",
						orderInfo.Order.Order.AppID, orderInfo.Order.Order.UserID, orderInfo.Order.Order.ID, orderIndex, usdAmount, i, commission)
				} else if commission.Start <= orderInfo.Order.Payment.CreateAt && commission.End == 0 {
					commissions[i].PayAmount += usdAmount
					commissions[i].CreateAt = orderInfo.Order.Payment.CreateAt
					logger.Sugar().Infof("app %v user %v order %v | %v commission %v commission %v container %v",
						orderInfo.Order.Order.AppID, orderInfo.Order.Order.UserID, orderInfo.Order.Order.ID, orderIndex, usdAmount, i, commission)
				}
			}
		} else {
			logger.Sugar().Infof("app %v user %v order %v | %v commission %v",
				orderInfo.Order.Order.AppID, orderInfo.Order.Order.UserID, orderInfo.Order.Order.ID, orderIndex, usdAmount)
			commissions = append(commissions, &npool.CommissionAmount{
				PayAmount: usdAmount,
				CreateAt:  orderInfo.Order.Payment.CreateAt,
			})
		}

		summary.Amount += usdAmount
		summarys[orderInfo.Good.Good.Good.CoinInfoID] = summary
	}

	username := inviteeResp.Info.User.EmailAddress
	if username == "" {
		username = inviteeResp.Info.User.PhoneNO
	}

	avatar := ""
	if inviteeResp.Info.Extra != nil {
		username = inviteeResp.Info.Extra.Username
		avatar = inviteeResp.Info.Extra.Avatar
	}

	logger.Sugar().Infof("app %v user %v commission %v summary %v", appID, inviteeID, commissions, summarys)

	return &npool.InvitationUserInfo{
		UserID:        inviteeResp.Info.User.ID,
		Username:      username,
		Avatar:        avatar,
		EmailAddress:  inviteeResp.Info.User.EmailAddress,
		Kol:           kol,
		MySummarys:    summarys,
		MyCommissions: commissions,
	}, myGoods, myCoins, nil
}

func getInvitations(appID, reqInviterID string, directOnly bool) (map[string]*npool.Invitation, *npool.InvitationUserInfo, error) { //nolint
	ctx := context.Background()

	_, err := grpc2.GetAppUserInfoByAppUser(ctx, &appusermgrpb.GetAppUserInfoByAppUserRequest{
		AppID:  appID,
		UserID: reqInviterID,
	})
	if err != nil {
		return nil, nil, xerrors.Errorf("fail get inviter %v user information: %v", reqInviterID, err)
	}

	goon := true
	invitations := map[string]*npool.Invitation{}
	invitations[reqInviterID] = &npool.Invitation{
		Invitees: []*npool.InvitationUserInfo{},
	}
	inviters := map[string]struct{}{}
	myGoods := map[string]*npool.Good{}
	myCoins := map[string]*coininfopb.CoinInfo{}
	myCounts := map[string]uint32{}

	inviterUserInfo, goods, coins, err := getInvitationUserInfo(appID, reqInviterID, myGoods, myCoins)
	if err != nil {
		return nil, nil, xerrors.Errorf("fail get inviter %v user info: %v", reqInviterID, err)
	}

	myGoods = goods
	myCoins = coins
	layer := 0

	// TODO: process deadloop
	for goon {
		goon = false

		for inviterID := range invitations { //nolint
			if _, ok := inviters[inviterID]; ok {
				continue
			}

			inviters[inviterID] = struct{}{}

			invitees, err := grpc2.GetRegistrationInvitationsByAppInviter(ctx, &inspirepb.GetRegistrationInvitationsByAppInviterRequest{
				AppID:     appID,
				InviterID: inviterID,
			})
			if err != nil {
				logger.Sugar().Errorf("fail get invitations by inviter %v: %v", inviterID, err)
				continue
			}

			myCounts[inviterID] = uint32(len(invitees))

			for i, info := range invitees {
				logger.Sugar().Infof("%v of %v layer %v user %v invited count %v", i, len(invitees), layer, inviterID, myCounts[inviterID])

				if info.AppID != appID || info.InviterID != inviterID {
					logger.Sugar().Errorf("invalid inviter id or app id")
					continue
				}

				userInfo, goods, coins, err := getInvitationUserInfo(appID, info.InviteeID, myGoods, myCoins)
				if err != nil {
					logger.Sugar().Errorf("fail get invitation user info: %v", err)
					continue
				}

				userInfo.JoinDate = info.CreateAt
				myGoods = goods
				myCoins = coins

				if _, ok := invitations[inviterID]; !ok {
					invitations[inviterID] = &npool.Invitation{
						Invitees: []*npool.InvitationUserInfo{},
					}
				}

				invitations[inviterID].Invitees = append(invitations[inviterID].Invitees, userInfo)

				if !directOnly || layer < 2 {
					if _, ok := invitations[userInfo.UserID]; !ok {
						invitations[userInfo.UserID] = &npool.Invitation{
							Invitees: []*npool.InvitationUserInfo{},
						}
					}
				}

				goon = true
			}
		}

		layer++
	}

	invitation := invitations[reqInviterID]
	inviterUserInfo.InvitedCount = uint32(len(invitation.Invitees))
	if inviterUserInfo.Summarys == nil {
		inviterUserInfo.Summarys = map[string]*npool.InvitationSummary{}
	}

	for _, comm := range inviterUserInfo.MyCommissions {
		inviterUserInfo.CommissionAmount += comm.PayAmount * float64(comm.Percent) / 100.0
		logger.Sugar().Infof("my own %v commission %v percent %v commission %v",
			inviterUserInfo, comm.PayAmount, comm.Percent, inviterUserInfo.CommissionAmount)
	}

	for _, invitee := range invitation.Invitees {
		curInviteeIDs := []string{invitee.UserID}
		foundInvitees := map[string]struct{}{}
		goon := true

		invitee.InvitedCount = myCounts[invitee.UserID]

		for coinID, summary := range invitee.MySummarys {
			if _, ok := inviterUserInfo.Summarys[coinID]; !ok {
				inviterUserInfo.Summarys[coinID] = &npool.InvitationSummary{}
			}
			mySummary := inviterUserInfo.Summarys[coinID]
			mySummary.Units += summary.Units
			mySummary.Amount += summary.Amount
			inviterUserInfo.Summarys[coinID] = mySummary
		}

		logger.Sugar().Infof("invitee %v commissions", invitee)

		for _, commission := range invitee.MyCommissions {
			for _, comm := range inviterUserInfo.MyCommissions {
				if comm.Amount != commission.Amount && commission.Percent > 0 {
					continue
				}
				if comm.Start <= commission.CreateAt && (commission.CreateAt < comm.End || comm.End == 0) {
					if comm.Percent <= commission.Percent {
						continue
					}
					inviterUserInfo.CommissionAmount += commission.PayAmount * float64(comm.Percent-commission.Percent) / 100.0
					invitee.CommissionAmount += commission.PayAmount * float64(commission.Percent) / 100.0
					logger.Sugar().Infof("invitee %v commission %v percent %v commission %v",
						invitee, comm.PayAmount, commission.Percent, inviterUserInfo.CommissionAmount)
					break
				}
			}
		}

		logger.Sugar().Infof("invitations %v", invitations)

		for goon {
			goon = false

			for _, curInviteeID := range curInviteeIDs {
				if _, ok := foundInvitees[curInviteeID]; ok {
					continue
				}

				foundInvitees[curInviteeID] = struct{}{}

				invitation, ok := invitations[curInviteeID]
				if !ok {
					continue
				}

				for _, iv := range invitation.Invitees {
					curInviteeIDs = append(curInviteeIDs, iv.UserID)

					logger.Sugar().Infof("caculate %v invited count %v", iv.UserID, myCounts[iv.UserID])
					iv.InvitedCount = myCounts[iv.UserID]

					if invitee.Summarys == nil {
						invitee.Summarys = map[string]*npool.InvitationSummary{}
					}

					for coinID, summary := range iv.MySummarys {
						if _, ok := invitee.Summarys[coinID]; !ok {
							invitee.Summarys[coinID] = &npool.InvitationSummary{}
						}
						// TODO: process different payment coin type
						mySummary := invitee.Summarys[coinID]
						mySummary.Units += summary.Units
						mySummary.Amount += summary.Amount
						invitee.Summarys[coinID] = mySummary

						if _, ok := inviterUserInfo.Summarys[coinID]; !ok {
							inviterUserInfo.Summarys[coinID] = &npool.InvitationSummary{}
						}
						mySummary = inviterUserInfo.Summarys[coinID]
						mySummary.Units += summary.Units
						mySummary.Amount += summary.Amount
						mySummary.Amount += summary.Amount
						inviterUserInfo.Summarys[coinID] = mySummary
					}

					for _, commission := range iv.MyCommissions {
						for _, comm := range invitee.MyCommissions {
							if comm.Amount != commission.Amount && commission.Percent > 0 {
								continue
							}
							if commission.Start > comm.Start || commission.End <= comm.End {
								continue
							}
							for _, comm1 := range inviterUserInfo.MyCommissions {
								if comm1.Amount != comm.Amount && commission.Percent > 0 && comm.Percent > 0 {
									continue
								}

								if comm1.Start <= commission.CreateAt && (commission.CreateAt < comm1.End || comm1.End == 0) {
									if comm1.Percent <= comm.Percent {
										continue
									}
									inviterUserInfo.CommissionAmount += commission.PayAmount * float64(comm1.Percent-comm.Percent) / 100.0
									logger.Sugar().Infof("invitee %v commission %v percent %v | %v commission %v",
										invitee, commission.PayAmount, comm1.Percent, comm.Percent, inviterUserInfo.CommissionAmount)
								}
							}
						}
					}

					goon = true
				}
			}
		}
	}

	if directOnly {
		return map[string]*npool.Invitation{
			reqInviterID: invitation,
		}, inviterUserInfo, nil
	}

	invitations[reqInviterID] = invitation

	return invitations, inviterUserInfo, nil
}