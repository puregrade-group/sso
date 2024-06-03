package profile

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/puregrade-group/sso/internal/domain/models"
)

type Profile struct {
	log             *slog.Logger
	profileProvider Provider
}

type Provider interface {
	SaveProfile(ctx context.Context, profile models.Profile) error
	GetProfile(ctx context.Context, profileId [16]byte) (models.Profile, error)
	DeleteProfile(ctx context.Context, profileId [16]byte) error
}

func New(
	log *slog.Logger,
	profileProvider Provider,
) *Profile {
	return &Profile{
		log:             log,
		profileProvider: profileProvider,
	}
}

func (p *Profile) Create(ctx context.Context,
	username string,
	avatarHash string,
	idp string,
	accountId string,
) (profileId [16]byte, err error) {
	const op = "profile.Create"

	log := p.log.With(
		slog.String("op", op),
		slog.String("idp", idp),
	)

	profileId = uuid.New()

	err = p.profileProvider.SaveProfile(
		ctx, models.Profile{
			Id:         profileId,
			Username:   username,
			AvatarHash: avatarHash,
			Idp:        idp,
			AccountId:  accountId,
		},
	)
	if err != nil {
		log.Error(
			"internal error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)
		return [16]byte{}, err
	}

	return profileId, err
}

func (p *Profile) Get(ctx context.Context,
	profileId [16]byte,
) (profile models.Profile, err error) {
	const op = "profile.Get"

	log := p.log.With(slog.String("op", op))

	id, err := uuid.FromBytes(profileId[:])
	if err != nil {
		log.Error(
			"wrong uuid", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)
		return models.Profile{}, err
	}

	log = p.log.With(slog.String("uuid", id.String()))

	profile, err = p.profileProvider.GetProfile(ctx, profileId)
	if err != nil {
		log.Error(
			"internal error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)
		return models.Profile{}, err
	}

	return profile, err
}

func (p *Profile) Delete(ctx context.Context,
	profileId [16]byte,
) (err error) {
	const op = "profile.Delete"

	log := p.log.With(slog.String("op", op))

	id, err := uuid.FromBytes(profileId[:])
	if err != nil {
		log.Error(
			"wrong uuid", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)
		return err
	}

	log = p.log.With(slog.String("uuid", id.String()))

	err = p.profileProvider.DeleteProfile(ctx, profileId)
	if err != nil {
		log.Error(
			"internal error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)
		return err
	}

	return nil
}
