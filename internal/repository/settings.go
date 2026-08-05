package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

// SettingsRepository manages system_settings key-value rows.
type SettingsRepository interface {
	ListAll(ctx context.Context) ([]model.SystemSettings, error)
	ListByGroup(ctx context.Context, group string) ([]model.SystemSettings, error)
	ListByGroups(ctx context.Context, groups []string) ([]model.SystemSettings, error)
	GetByKey(ctx context.Context, key string) (*model.SystemSettings, error)
	GetByKeys(ctx context.Context, keys []string) (map[string]model.SystemSettings, error)
	GetByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error)
	GetByGroups(ctx context.Context, groups []string) (map[string]model.SystemSettings, error)
	// Upsert sets value for key; creates row when missing.
	Upsert(ctx context.Context, key, value string, settingType uint8, description string) error
	// UpsertMany upserts multiple keys in a transaction.
	UpsertMany(ctx context.Context, items []SettingUpsert) error
}

// SettingUpsert is one upsert payload.
type SettingUpsert struct {
	Key         string
	Group       string
	Value       string
	SettingType uint8
	Description string
}

type settingsRepo struct {
	db *database.DB
}

// NewSettingsRepo creates a SettingsRepository.
func NewSettingsRepo(db *database.DB) SettingsRepository {
	return &settingsRepo{db: db}
}

func (r *settingsRepo) ListAll(ctx context.Context) ([]model.SystemSettings, error) {
	var items []model.SystemSettings
	err := r.db.NewSelect().Model(&items).Order("setting_key ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	return items, nil
}

func (r *settingsRepo) ListByGroup(ctx context.Context, group string) ([]model.SystemSettings, error) {
	var items []model.SystemSettings
	err := r.db.NewSelect().Model(&items).Where("setting_group = ?", group).Order("setting_key ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list settings by group %s: %w", group, err)
	}
	return items, nil
}

func (r *settingsRepo) ListByGroups(ctx context.Context, groups []string) ([]model.SystemSettings, error) {
	if len(groups) == 0 {
		return nil, nil
	}
	var items []model.SystemSettings
	err := r.db.NewSelect().Model(&items).Where("setting_group IN (?)", bun.List(groups)).Order("setting_key ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list settings by groups: %w", err)
	}
	return items, nil
}

func (r *settingsRepo) GetByKey(ctx context.Context, key string) (*model.SystemSettings, error) {
	item := new(model.SystemSettings)
	err := r.db.NewSelect().Model(item).Where("setting_key = ?", key).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get setting %s: %w", key, err)
	}
	return item, nil
}

func (r *settingsRepo) GetByKeys(ctx context.Context, keys []string) (map[string]model.SystemSettings, error) {
	out := make(map[string]model.SystemSettings, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	var items []model.SystemSettings
	err := r.db.NewSelect().Model(&items).Where("setting_key IN (?)", bun.List(keys)).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings by keys: %w", err)
	}
	for _, it := range items {
		out[it.SettingKey] = it
	}
	return out, nil
}

func (r *settingsRepo) GetByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error) {
	items, err := r.ListByGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	out := make(map[string]model.SystemSettings, len(items))
	for _, it := range items {
		out[it.SettingKey] = it
	}
	return out, nil
}

func (r *settingsRepo) GetByGroups(ctx context.Context, groups []string) (map[string]model.SystemSettings, error) {
	items, err := r.ListByGroups(ctx, groups)
	if err != nil {
		return nil, err
	}
	out := make(map[string]model.SystemSettings, len(items))
	for _, it := range items {
		out[it.SettingKey] = it
	}
	return out, nil
}

func (r *settingsRepo) Upsert(ctx context.Context, key, value string, settingType uint8, description string) error {
	return r.UpsertMany(ctx, []SettingUpsert{{
		Key: key, Value: value, SettingType: settingType, Description: description,
	}})
}

func (r *settingsRepo) UpsertMany(ctx context.Context, items []SettingUpsert) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, it := range items {
			existing := new(model.SystemSettings)
			err := tx.NewSelect().Model(existing).Where("setting_key = ?", it.Key).Scan(ctx)
			if errors.Is(err, sql.ErrNoRows) {
				val := it.Value
				row := &model.SystemSettings{
					SettingKey:   it.Key,
					SettingGroup: it.Group,
					SettingValue: val,
					SettingType:  it.SettingType,
					Description:  it.Description,
				}
				if _, insertErr := tx.NewInsert().Model(row).Exec(ctx); insertErr != nil {
					return fmt.Errorf("insert setting %s: %w", it.Key, insertErr)
				}
				continue
			}
			if err != nil {
				return fmt.Errorf("select setting %s: %w", it.Key, err)
			}
			val := it.Value
			existing.SettingValue = val
			if it.SettingType > 0 {
				existing.SettingType = it.SettingType
			}
			if it.Description != "" {
				existing.Description = it.Description
			}
			if it.Group != "" {
				existing.SettingGroup = it.Group
			}
			if _, err := tx.NewUpdate().Model(existing).
				Column("setting_value", "setting_type", "description", "setting_group", "updated_at").
				WherePK().
				Exec(ctx); err != nil {
				return fmt.Errorf("update setting %s: %w", it.Key, err)
			}
		}
		return nil
	})
}
