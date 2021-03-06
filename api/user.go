//go:build !codeanalysis
// +build !codeanalysis

package api

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/cloud-hashing-apis"

	"github.com/NpoolPlatform/cloud-hashing-apis/pkg/middleware/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Signup(ctx context.Context, in *npool.SignupRequest) (*npool.SignupResponse, error) {
	resp, err := user.Signup(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("sign up error: %w", err)
		return &npool.SignupResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) UpdatePassword(ctx context.Context, in *npool.UpdatePasswordRequest) (*npool.UpdatePasswordResponse, error) {
	resp, err := user.UpdatePassword(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("update password error: %w", err)
		return &npool.UpdatePasswordResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) UpdatePasswordByAppUser(ctx context.Context, in *npool.UpdatePasswordByAppUserRequest) (*npool.UpdatePasswordByAppUserResponse, error) {
	resp, err := user.UpdatePasswordByAppUser(ctx, in, true)
	if err != nil {
		logger.Sugar().Errorf("update password by app user error: %w", err)
		return &npool.UpdatePasswordByAppUserResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) UpdateEmailAddress(ctx context.Context, in *npool.UpdateEmailAddressRequest) (*npool.UpdateEmailAddressResponse, error) {
	resp, err := user.UpdateEmailAddress(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("update email address error: %w", err)
		return &npool.UpdateEmailAddressResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) UpdatePhoneNO(ctx context.Context, in *npool.UpdatePhoneNORequest) (*npool.UpdatePhoneNOResponse, error) {
	resp, err := user.UpdatePhoneNO(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("update phone NO error: %w", err)
		return &npool.UpdatePhoneNOResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) UpdateAccount(ctx context.Context, in *npool.UpdateAccountRequest) (*npool.UpdateAccountResponse, error) {
	resp, err := user.UpdateAccount(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("update account error: %v", err)
		return &npool.UpdateAccountResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) UpdateAppUserExtra(ctx context.Context, in *npool.UpdateAppUserExtraRequest) (*npool.UpdateAppUserExtraResponse, error) {
	resp, err := user.UpdateAppUserExtra(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("update app user extra error: %v", err)
		return &npool.UpdateAppUserExtraResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) CreateAppUserExtra(ctx context.Context, in *npool.CreateAppUserExtraRequest) (*npool.CreateAppUserExtraResponse, error) {
	resp, err := user.CreateAppUserExtra(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("create app user extra error: %v", err)
		return &npool.CreateAppUserExtraResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) UpdateAppUserControl(ctx context.Context, in *npool.UpdateAppUserControlRequest) (*npool.UpdateAppUserControlResponse, error) {
	resp, err := user.UpdateAppUserControl(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("update app user control error: %v", err)
		return &npool.UpdateAppUserControlResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *Server) CreateAppUserControl(ctx context.Context, in *npool.CreateAppUserControlRequest) (*npool.CreateAppUserControlResponse, error) {
	resp, err := user.CreateAppUserControl(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("create app user control error: %v", err)
		return &npool.CreateAppUserControlResponse{}, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
