package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/user"
)

type mysqlUserRepository struct{ client *ent.Client }

func NewUserRepository(client *ent.Client) UserRepository {
	return &mysqlUserRepository{client: client}
}

func (r *mysqlUserRepository) Create(ctx context.Context, input CreateUserInput) (*ent.User, error) {
	now := time.Now().Unix()
	builder := r.client.User.Create().
		SetUsername(input.Username).
		SetPassword(input.Password).
		SetSalt(input.Salt).
		SetNickname(input.Nickname).
		SetEmail(input.Email).
		SetMobile(input.Mobile).
		SetJoinip(input.JoinIP).
		SetJointime(now).
		SetCreatetime(now).
		SetUpdatetime(now)
	return builder.Save(ctx)
}

func (r *mysqlUserRepository) GetByID(ctx context.Context, id int64) (*ent.User, error) {
	return r.client.User.Get(ctx, int(id))
}

func (r *mysqlUserRepository) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.client.User.Query().Where(user.UsernameEQ(username)).Only(ctx)
}

func (r *mysqlUserRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
}

func (r *mysqlUserRepository) GetByMobile(ctx context.Context, mobile string) (*ent.User, error) {
	return r.client.User.Query().Where(user.MobileEQ(mobile)).Only(ctx)
}

func (r *mysqlUserRepository) GetByToken(ctx context.Context, token string) (*ent.User, error) {
	return r.client.User.Query().Where(user.TokenEQ(token), user.StatusEQ("normal")).Only(ctx)
}

func (r *mysqlUserRepository) Update(ctx context.Context, id int64, input UpdateUserInput) (*ent.User, error) {
	upd := r.client.User.UpdateOneID(int(id)).SetUpdatetime(time.Now().Unix())
	if input.Nickname != nil {
		upd = upd.SetNickname(*input.Nickname)
	}
	if input.Avatar != nil {
		upd = upd.SetAvatar(*input.Avatar)
	}
	if input.Email != nil {
		upd = upd.SetEmail(*input.Email)
	}
	if input.Mobile != nil {
		upd = upd.SetMobile(*input.Mobile)
	}
	if input.Bio != nil {
		upd = upd.SetBio(*input.Bio)
	}
	if input.Birthday != nil {
		upd = upd.SetBirthday(*input.Birthday)
	}
	if input.Gender != nil {
		upd = upd.SetGender(*input.Gender)
	}
	return upd.Save(ctx)
}

func (r *mysqlUserRepository) UpdatePassword(ctx context.Context, id int64, hashedPassword, salt string) error {
	_, err := r.client.User.UpdateOneID(int(id)).
		SetPassword(hashedPassword).
		SetSalt(salt).
		SetUpdatetime(time.Now().Unix()).
		Save(ctx)
	return err
}

func (r *mysqlUserRepository) UpdateToken(ctx context.Context, id int64, token string) error {
	_, err := r.client.User.UpdateOneID(int(id)).
		SetToken(token).
		SetUpdatetime(time.Now().Unix()).
		Save(ctx)
	return err
}

func (r *mysqlUserRepository) UpdateLoginInfo(ctx context.Context, id int64, loginIP string, loginTime int64) error {
	_, err := r.client.User.UpdateOneID(int(id)).
		SetPrevtime(r.mustGetPrevTime(ctx, id)).
		SetLogintime(loginTime).
		SetLoginip(loginIP).
		SetLoginfailure(0).
		SetUpdatetime(time.Now().Unix()).
		Save(ctx)
	return err
}

// mustGetPrevTime returns the current logintime so it can become the new prevtime.
// If the lookup fails we fall back to 0.
func (r *mysqlUserRepository) mustGetPrevTime(ctx context.Context, id int64) int64 {
	u, err := r.client.User.Get(ctx, int(id))
	if err != nil {
		return 0
	}
	return u.Logintime
}
