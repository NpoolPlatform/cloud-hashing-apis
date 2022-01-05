package user

import (
	"context"
	"sync"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	"github.com/NpoolPlatform/cloud-hashing-apis/message/npool"

	grpc2 "github.com/NpoolPlatform/cloud-hashing-apis/pkg/grpc"
	order "github.com/NpoolPlatform/cloud-hashing-apis/pkg/middleware/order"

	inspirepb "github.com/NpoolPlatform/cloud-hashing-inspire/message/npool"
	orderconst "github.com/NpoolPlatform/cloud-hashing-order/pkg/const"
	coininfopb "github.com/NpoolPlatform/message/npool/coininfo"
	usermgrpb "github.com/NpoolPlatform/user-management/message/npool"

	"golang.org/x/xerrors"
)

var (
	appInvitations = map[string]map[string]map[string]*npool.Invitation{}
	mutex          = sync.Mutex{}
)

func addWatcher(appID, inviterID string) {
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
}

func Run() {
	go func() {
		ticker := time.NewTicker(4 * time.Hour)

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

		for appID, inviters := range appInviters {
			for _, inviterID := range inviters {
				invitations, err := getInvitations(appID, inviterID, false, false)
				if err != nil {
					logger.Sugar().Errorf("fail get invitations: %v", err)
					continue
				}

				mutex.Lock()
				appInvitations[appID][inviterID] = invitations
				mutex.Unlock()
			}
		}

		<-ticker.C
	}()
}

func getFullInvitations(appID, inviterID string) (map[string]*npool.Invitation, error) {
	mutex.Lock()
	invitations := appInvitations[appID][inviterID]
	mutex.Unlock()

	if len(invitations) > 0 {
		return invitations, nil
	}

	invitations, err := getInvitations(appID, inviterID, false, true)
	if err != nil {
		return nil, xerrors.Errorf("fail get invitations: %v", err)
	}

	return invitations, nil
}

func getDirectInvitations(appID, inviterID string) (map[string]*npool.Invitation, error) {
	mutex.Lock()
	invitations := appInvitations[appID][inviterID]
	mutex.Unlock()

	if len(invitations) > 0 {
		return invitations, nil
	}

	invitations, err := getInvitations(appID, inviterID, true, true)
	if err != nil {
		return nil, xerrors.Errorf("fail get invitations: %v", err)
	}

	return invitations, nil
}

func getInvitations(appID, reqInviterID string, directOnly bool, noOrder bool) (map[string]*npool.Invitation, error) { //nolint
	ctx := context.Background()

	_, err := grpc2.GetUser(ctx, &usermgrpb.GetUserRequest{
		AppID:  appID,
		UserID: reqInviterID,
	})
	if err != nil {
		return nil, xerrors.Errorf("fail get inviter %v user information: %v", reqInviterID, err)
	}

	goon := true
	invitations := map[string]*npool.Invitation{}
	invitations[reqInviterID] = &npool.Invitation{
		Invitees: []*npool.InvitationUserInfo{},
	}
	inviters := map[string]struct{}{}
	myGoods := map[string]*npool.GoodDetail{}
	myCoins := map[string]*coininfopb.CoinInfo{}
	myCounts := map[string]uint32{}

	// TODO: process deadloop
	for goon {
		goon = false

		for inviterID := range invitations { //nolint
			if _, ok := inviters[inviterID]; ok {
				continue
			}

			inviters[inviterID] = struct{}{}

			resp, err := grpc2.GetRegistrationInvitationsByAppInviter(ctx, &inspirepb.GetRegistrationInvitationsByAppInviterRequest{
				AppID:     appID,
				InviterID: inviterID,
			})
			if err != nil {
				logger.Sugar().Errorf("fail get invitations by inviter %v: %v", inviterID, err)
				continue
			}

			myCounts[inviterID] = uint32(len(resp.Infos))

			for i, info := range resp.Infos {
				logger.Sugar().Infof("%v of %v", i, len(resp.Infos))

				if info.AppID != appID || info.InviterID != inviterID {
					logger.Sugar().Errorf("invalid inviter id or app id")
					continue
				}

				inviteeResp, err := grpc2.GetUser(ctx, &usermgrpb.GetUserRequest{
					AppID:  appID,
					UserID: info.InviteeID,
				})
				if err != nil {
					logger.Sugar().Errorf("fail get invitee %v user info: %v", info.InviteeID, err)
					continue
				}

				resp1, err := grpc2.GetUserInvitationCodeByAppUser(ctx, &inspirepb.GetUserInvitationCodeByAppUserRequest{
					AppID:  appID,
					UserID: inviteeResp.Info.UserID,
				})
				if err != nil {
					logger.Sugar().Errorf("fail get user invitation code: %v", err)
					continue
				}

				summarys := map[string]*npool.InvitationSummary{}

				if !noOrder {
					resp2, goods, coins, err := order.GetOrdersShortDetailByAppUser(ctx, &npool.GetOrdersDetailByAppUserRequest{
						AppID:  appID,
						UserID: inviteeResp.Info.UserID,
					}, myGoods, myCoins)
					if err != nil {
						logger.Sugar().Errorf("fail get orders detail by app user: %v", err)
						continue
					}

					myGoods = goods
					myCoins = coins

					for _, orderInfo := range resp2.Details {
						if orderInfo.Payment == nil {
							continue
						}

						if orderInfo.Payment.State != orderconst.PaymentStateDone {
							continue
						}

						if _, ok := summarys[orderInfo.Good.CoinInfo.ID]; !ok {
							summarys[orderInfo.Good.CoinInfo.ID] = &npool.InvitationSummary{}
						}

						summary := summarys[orderInfo.Good.CoinInfo.ID]
						summary.Units += orderInfo.Units
						summary.Amount += orderInfo.Payment.Amount
						summarys[orderInfo.Good.CoinInfo.ID] = summary
					}
				}

				kol := false
				if resp1.Info != nil {
					kol = true
				}

				if _, ok := invitations[inviterID]; !ok {
					invitations[inviterID] = &npool.Invitation{
						Invitees: []*npool.InvitationUserInfo{},
					}
				}

				invitations[inviterID].Invitees = append(
					invitations[inviterID].Invitees, &npool.InvitationUserInfo{
						UserID:       inviteeResp.Info.UserID,
						Username:     inviteeResp.Info.Username,
						Avatar:       inviteeResp.Info.Avatar,
						EmailAddress: inviteeResp.Info.EmailAddress,
						Kol:          kol,
						MySummarys:   summarys,
					})

				if !directOnly {
					if _, ok := invitations[inviteeResp.Info.UserID]; !ok {
						invitations[inviteeResp.Info.UserID] = &npool.Invitation{
							Invitees: []*npool.InvitationUserInfo{},
						}
					}
				}

				goon = true
			}
		}
	}

	invitation := invitations[reqInviterID]

	for _, invitee := range invitation.Invitees {
		curInviteeIDs := []string{invitee.UserID}
		foundInvitees := map[string]struct{}{}
		goon := true

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

					logger.Sugar().Infof("start caculate %v", iv.UserID)
					iv.InvitedCount = myCounts[iv.UserID]

					for coinID, summary := range iv.MySummarys {
						if _, ok := invitee.Summarys[coinID]; !ok {
							invitee.Summarys[coinID] = &npool.InvitationSummary{}
						}
						// TODO: process different payment coin type
						mySummary := invitee.Summarys[coinID]
						mySummary.Units += summary.Units
						mySummary.Amount += summary.Amount
						invitee.Summarys[coinID] = mySummary
					}

					logger.Sugar().Infof("end caculate %v", iv.UserID)
					goon = true
				}
			}
		}
	}

	if directOnly {
		return map[string]*npool.Invitation{
			reqInviterID: invitation,
		}, nil
	}

	invitations[reqInviterID] = invitation

	return invitations, nil
}
