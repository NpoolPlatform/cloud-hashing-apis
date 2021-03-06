//go:build !codeanalysis
// +build !codeanalysis

package api

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/cloud-hashing-apis"

	"github.com/NpoolPlatform/cloud-hashing-apis/pkg/middleware/referral"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetReferrals(ctx context.Context, in *npool.GetReferralsRequest) (*npool.GetReferralsResponse, error) {
	resp, err := referral.GetReferrals(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("get referrals error: %w", err)
		return &npool.GetReferralsResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

// 管理员根据用户ID拉取
func (s *Server) GetUserReferrals(ctx context.Context, in *npool.GetUserReferralsRequest) (*npool.GetUserReferralsResponse, error) {
	resp, err := referral.GetUserReferrals(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("get referrals error: %w", err)
		return &npool.GetUserReferralsResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) GetLayeredReferrals(ctx context.Context, in *npool.GetLayeredReferralsRequest) (*npool.GetLayeredReferralsResponse, error) {
	resp, err := referral.GetLayeredReferrals(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("get layered referrals error: %w", err)
		return &npool.GetLayeredReferralsResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) CreateInvitationCode(ctx context.Context, in *npool.CreateInvitationCodeRequest) (*npool.CreateInvitationCodeResponse, error) {
	code, err := referral.CreateInvitationCode(
		ctx,
		in.GetAppID(), in.GetUserID(), in.GetTargetUserID(), in.GetLangID(),
		in.GetInviterName(), in.GetInviteeName(),
		in.GetInfo())
	if err != nil {
		logger.Sugar().Errorf("create invitation code error: %w", err)
		return &npool.CreateInvitationCodeResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &npool.CreateInvitationCodeResponse{
		Info: code,
	}, nil
}
