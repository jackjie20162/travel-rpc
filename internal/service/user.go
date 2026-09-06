package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/auth"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/travel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	travel.UnimplementedUserServiceServer
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

// ── Register ──

func (s *UserService) Register(ctx context.Context, req *travel.RegisterRequest) (*travel.AuthToken, error) {
	if req == nil || req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	// Check if username already exists
	if existing, _ := s.users.GetByUsername(ctx, req.GetUsername()); existing != nil {
		return nil, status.Error(codes.AlreadyExists, "username already taken")
	}

	// Check mobile uniqueness if provided
	if req.GetMobile() != "" {
		if existing, _ := s.users.GetByMobile(ctx, req.GetMobile()); existing != nil {
			return nil, status.Error(codes.AlreadyExists, "mobile already registered")
		}
	}

	salt := generateSalt()
	hashedPwd := hashPassword(req.GetPassword(), salt)
	nickname := req.GetNickname()
	if nickname == "" {
		nickname = req.GetUsername()
	}

	u, err := s.users.Create(ctx, repository.CreateUserInput{
		Username: req.GetUsername(),
		Password: hashedPwd,
		Salt:     salt,
		Email:    req.GetEmail(),
		Mobile:   req.GetMobile(),
		Nickname: nickname,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	token := generateToken()
	if err := s.users.UpdateToken(ctx, int64(u.ID), token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	u.Token = token

	return &travel.AuthToken{Token: token, User: toUser(u)}, nil
}

// ── Login ──

func (s *UserService) Login(ctx context.Context, req *travel.LoginRequest) (*travel.AuthToken, error) {
	if req == nil || req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	u, err := s.users.GetByUsername(ctx, req.GetUsername())
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if u.Status != "normal" {
		return nil, status.Error(codes.PermissionDenied, "account is disabled")
	}

	if hashPassword(req.GetPassword(), u.Salt) != u.Password {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	token := generateToken()
	if err := s.users.UpdateToken(ctx, int64(u.ID), token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_ = s.users.UpdateLoginInfo(ctx, int64(u.ID), "", time.Now().Unix())
	u.Token = token

	return &travel.AuthToken{Token: token, User: toUser(u)}, nil
}

// ── LoginByMobile ──

func (s *UserService) LoginByMobile(ctx context.Context, req *travel.LoginByMobileRequest) (*travel.AuthToken, error) {
	if req == nil || req.GetMobile() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "mobile and password are required")
	}

	u, err := s.users.GetByMobile(ctx, req.GetMobile())
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if u.Status != "normal" {
		return nil, status.Error(codes.PermissionDenied, "account is disabled")
	}

	if hashPassword(req.GetPassword(), u.Salt) != u.Password {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	token := generateToken()
	if err := s.users.UpdateToken(ctx, int64(u.ID), token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_ = s.users.UpdateLoginInfo(ctx, int64(u.ID), "", time.Now().Unix())
	u.Token = token

	return &travel.AuthToken{Token: token, User: toUser(u)}, nil
}

// ── GetProfile ──

func (s *UserService) GetProfile(ctx context.Context, req *travel.UserIdRequest) (*travel.User, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	u, err := s.users.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return toUser(u), nil
}

// ── UpdateProfile ──

func (s *UserService) UpdateProfile(ctx context.Context, req *travel.UpdateProfileRequest) (*travel.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// The caller (travel-api) must extract user ID from the authenticated token
	// and pass it via gRPC metadata. For now we use a simple approach:
	// the user ID comes from the authenticated context.
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	input := repository.UpdateUserInput{}
	if req.GetNickname() != "" {
		input.Nickname = strPtr(req.GetNickname())
	}
	if req.GetAvatar() != "" {
		input.Avatar = strPtr(req.GetAvatar())
	}
	if req.GetEmail() != "" {
		input.Email = strPtr(req.GetEmail())
	}
	if req.GetMobile() != "" {
		input.Mobile = strPtr(req.GetMobile())
	}
	if req.GetBio() != "" {
		input.Bio = strPtr(req.GetBio())
	}
	if req.GetBirthday() != "" {
		input.Birthday = strPtr(req.GetBirthday())
	}
	if req.GetGender() != 0 {
		g := int8(req.GetGender())
		input.Gender = &g
	}

	u, err := s.users.Update(ctx, userID, input)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toUser(u), nil
}

// ── ChangePassword ──

func (s *UserService) ChangePassword(ctx context.Context, req *travel.ChangePasswordRequest) (*travel.OkResponse, error) {
	if req == nil || req.GetOldPassword() == "" || req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "old and new password are required")
	}

	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if hashPassword(req.GetOldPassword(), u.Salt) != u.Password {
		return nil, status.Error(codes.Unauthenticated, "invalid old password")
	}

	newSalt := generateSalt()
	newHash := hashPassword(req.GetNewPassword(), newSalt)
	if err := s.users.UpdatePassword(ctx, userID, newHash, newSalt); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Invalidate existing token so user must re-login
	_ = s.users.UpdateToken(ctx, userID, "")

	return &travel.OkResponse{Ok: true, Msg: "password changed"}, nil
}

// ── Helpers ──

func toUser(u *ent.User) *travel.User {
	if u == nil {
		return nil
	}
	p := &travel.User{
		Id:       int64(u.ID),
		Username: u.Username,
		Nickname: u.Nickname,
		Email:    u.Email,
		Mobile:   u.Mobile,
		Avatar:   u.Avatar,
		Level:    int32(u.Level),
		Gender:   int32(u.Gender),
		Bio:      u.Bio,
		Money:    u.Money,
		Score:    int32(u.Score),
		Jointime:  derefInt64(u.Jointime),
		Logintime: u.Logintime,
		Loginip:  u.Loginip,
		Status:   u.Status,
		Createtime: derefInt64(u.Createtime),
	}
	if u.Birthday != nil {
		p.Birthday = *u.Birthday
	}
	return p
}

func hashPassword(password, salt string) string {
	h := sha256.Sum256([]byte(password))
	combined := hex.EncodeToString(h[:]) + salt
	result := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(result[:])
}

func generateSalt() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func strPtr(s string) *string { return &s }

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func authenticatedUserID(ctx context.Context) (int64, error) {
	return auth.AuthenticatedUserID(ctx)
}
