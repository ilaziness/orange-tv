package service

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapToStorageSettingsMasksSecret(t *testing.T) {
	t.Parallel()
	m := map[string]model.SystemSettings{
		constant.SettingStorageProvider: {
			SettingKey:   constant.SettingStorageProvider,
			SettingValue: "aliyun",
		},
		constant.SettingStorageAliyun: {
			SettingKey: constant.SettingStorageAliyun,
			SettingValue: MarshalStorageProviderRaw(StorageProviderRaw{
				Bucket:    "b1",
				AccessKey: "ABCDEFGH",
				SecretKey: "super-secret",
				CDNDomain: "https://cdn.example.com",
			}),
		},
	}
	out := MapToStorageSettings(m)
	assert.Equal(t, "aliyun", out.Provider)
	assert.Equal(t, "b1", out.Aliyun.Bucket)
	assert.True(t, out.Aliyun.AccessConfigured)
	assert.True(t, out.Aliyun.SecretConfigured)
	assert.Empty(t, out.Aliyun.AccessKey)
	assert.Empty(t, out.Aliyun.SecretKey)
}

func TestValidateStorageProviderConfigRequiresCDN(t *testing.T) {
	t.Parallel()
	err := ValidateStorageProviderConfig(constant.StorageProviderAliyun, StorageProviderRaw{
		Bucket:    "b",
		AccessKey: "ak",
		SecretKey: "sk",
		Region:    "oss-cn-hangzhou",
	}, true)
	require.Error(t, err)

	err = ValidateStorageProviderConfig(constant.StorageProviderAliyun, StorageProviderRaw{
		Bucket:    "b",
		AccessKey: "ak",
		SecretKey: "sk",
		Region:    "oss-cn-hangzhou",
		CDNDomain: "https://cdn.example.com",
	}, true)
	require.NoError(t, err)
}

func TestValidateStorageProviderConfigRequiresRegion(t *testing.T) {
	t.Parallel()
	err := ValidateStorageProviderConfig(constant.StorageProviderAliyun, StorageProviderRaw{
		Bucket: "b", AccessKey: "ak", SecretKey: "sk", CDNDomain: "https://cdn.example.com",
	}, true)
	require.Error(t, err)

	err = ValidateStorageProviderConfig(constant.StorageProviderAliyun, StorageProviderRaw{
		Bucket: "b", AccessKey: "ak", SecretKey: "sk", Endpoint: "oss-cn-hangzhou.aliyuncs.com",
		CDNDomain: "https://cdn.example.com",
	}, true)
	require.NoError(t, err)

	err = ValidateStorageProviderConfig(constant.StorageProviderTencent, StorageProviderRaw{
		Bucket: "b", AccessKey: "ak", SecretKey: "sk", CDNDomain: "https://cdn.example.com",
	}, true)
	require.Error(t, err)

	err = ValidateStorageProviderConfig(constant.StorageProviderTencent, StorageProviderRaw{
		Bucket: "b", AccessKey: "ak", SecretKey: "sk", Region: "ap-guangzhou",
		CDNDomain: "https://cdn.example.com",
	}, true)
	require.NoError(t, err)

	err = ValidateStorageProviderConfig(constant.StorageProviderQiniu, StorageProviderRaw{
		Bucket: "b", AccessKey: "ak", SecretKey: "sk", CDNDomain: "https://cdn.example.com",
	}, true)
	require.NoError(t, err)
}
