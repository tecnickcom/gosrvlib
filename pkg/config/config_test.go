// +build unit

package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// nolint:maligned
type testData struct {
	Str string `mapstructure:"str"`
	Int int    `mapstructure:"int"`
	Arr []int  `mapstructure:"arr"`
}

// nolint:maligned
type testConfig struct {
	validateErr error
	BaseConfig  `mapstructure:",squash"`
	Boolean     bool                         `mapstructure:"bool"`
	String      string                       `mapstructure:"string"`
	Int         int                          `mapstructure:"int"`
	Int8        int8                         `mapstructure:"int8"`
	Int16       int16                        `mapstructure:"int16"`
	Int32       int32                        `mapstructure:"int32"`
	Int64       int64                        `mapstructure:"int64"`
	UInt        uint                         `mapstructure:"uint"`
	UInt8       uint8                        `mapstructure:"uint8"`
	UInt16      uint16                       `mapstructure:"uint16"`
	UInt32      uint32                       `mapstructure:"uint32"`
	UInt64      uint64                       `mapstructure:"uint64"`
	Float32     float32                      `mapstructure:"float32"`
	Float64     float32                      `mapstructure:"float64"`
	Slice       []string                     `mapstructure:"slice"`
	MapString   map[string]string            `mapstructure:"mapstring"`
	NestedMap   map[string]map[string]string `mapstructure:"nestedmap"`
	Data        testData                     `mapstructure:"data"`
	MapData     map[string]testData          `mapstructure:"mapdata"`
}

func (tc *testConfig) SetDefaults(v Viper) {
	v.SetDefault("drink", "c0ffee")
}

func (tc *testConfig) Validate() error {
	return tc.validateErr
}

func Test_configureConfigSearchPath(t *testing.T) {
	tests := []struct {
		name       string
		configDir  string
		setupMocks func(m *MockViper)
	}{
		{
			name: "without config dir",
			setupMocks: func(m *MockViper) {
				m.EXPECT().AddConfigPath("./")
				m.EXPECT().AddConfigPath("$HOME/.test/")
				m.EXPECT().AddConfigPath("/etc/test/")
			},
		},
		{
			name:      "with config dir",
			configDir: "/config_source_test/",
			setupMocks: func(m *MockViper) {
				m.EXPECT().AddConfigPath("/config_source_test/")
				m.EXPECT().AddConfigPath("./")
				m.EXPECT().AddConfigPath("$HOME/.test/")
				m.EXPECT().AddConfigPath("/etc/test/")
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tmv := NewMockViper(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(tmv)
			}

			configureSearchPath(tmv, "test", tt.configDir)
		})
	}
}

func Test_loadLocalConfig(t *testing.T) {
	tests := []struct {
		name           string
		configContent  []byte
		envDataContent []byte
		setupViper     func(ctrl *gomock.Controller) Viper
		want           *remoteSourceConfig
		wantErr        bool
	}{
		{
			name: "fails with read config error",
			setupViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyRemoteConfigProvider, defaultRemoteConfigProvider)
				mock.EXPECT().SetDefault(keyRemoteConfigEndpoint, defaultRemoteConfigEndpoint)
				mock.EXPECT().SetDefault(keyRemoteConfigPath, defaultRemoteConfigPath)
				mock.EXPECT().SetDefault(keyRemoteConfigSecretKeyring, defaultRemoteConfigSecretKeyring)
				mock.EXPECT().SetDefault(keyLogFormat, defaultLogFormat)
				mock.EXPECT().SetDefault(keyLogLevel, defaultLogLevel)
				mock.EXPECT().SetDefault(keyLogAddress, defaultLogAddress)
				mock.EXPECT().SetDefault(keyLogNetwork, defaultLogNetwork)
				mock.EXPECT().SetDefault("drink", "c0ffee")
				mock.EXPECT().SetConfigName(defaultConfigName)
				mock.EXPECT().SetConfigType(defaultConfigType)
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().AutomaticEnv()
				mock.EXPECT().SetEnvPrefix("test")
				mock.EXPECT().BindEnv(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig().Return(fmt.Errorf("read config error"))
				return mock
			},
			wantErr: true,
		},
		{
			name: "fails with unmarshal error",
			setupViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyRemoteConfigProvider, defaultRemoteConfigProvider)
				mock.EXPECT().SetDefault(keyRemoteConfigEndpoint, defaultRemoteConfigEndpoint)
				mock.EXPECT().SetDefault(keyRemoteConfigPath, defaultRemoteConfigPath)
				mock.EXPECT().SetDefault(keyRemoteConfigSecretKeyring, defaultRemoteConfigSecretKeyring)
				mock.EXPECT().SetDefault(keyLogFormat, defaultLogFormat)
				mock.EXPECT().SetDefault(keyLogLevel, defaultLogLevel)
				mock.EXPECT().SetDefault(keyLogAddress, defaultLogAddress)
				mock.EXPECT().SetDefault(keyLogNetwork, defaultLogNetwork)
				mock.EXPECT().SetDefault("drink", "c0ffee")
				mock.EXPECT().SetConfigName(defaultConfigName)
				mock.EXPECT().SetConfigType(defaultConfigType)
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().AutomaticEnv()
				mock.EXPECT().SetEnvPrefix("test")
				mock.EXPECT().BindEnv(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig()
				mock.EXPECT().Unmarshal(gomock.Any()).Return(fmt.Errorf("unmarshal error"))
				return mock
			},
			wantErr: true,
		},
		{
			name: "succeed",
			setupViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyRemoteConfigProvider, defaultRemoteConfigProvider)
				mock.EXPECT().SetDefault(keyRemoteConfigEndpoint, defaultRemoteConfigEndpoint)
				mock.EXPECT().SetDefault(keyRemoteConfigPath, defaultRemoteConfigPath)
				mock.EXPECT().SetDefault(keyRemoteConfigSecretKeyring, defaultRemoteConfigSecretKeyring)
				mock.EXPECT().SetDefault(keyLogFormat, defaultLogFormat)
				mock.EXPECT().SetDefault(keyLogLevel, defaultLogLevel)
				mock.EXPECT().SetDefault(keyLogAddress, defaultLogAddress)
				mock.EXPECT().SetDefault(keyLogNetwork, defaultLogNetwork)
				mock.EXPECT().SetDefault("drink", "c0ffee")
				mock.EXPECT().SetConfigName(defaultConfigName)
				mock.EXPECT().SetConfigType(defaultConfigType)
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().AutomaticEnv()
				mock.EXPECT().SetEnvPrefix("test")
				mock.EXPECT().BindEnv(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig()
				mock.EXPECT().Unmarshal(gomock.Any())
				return mock
			},
			want:    &remoteSourceConfig{},
			wantErr: false,
		},
		{
			name:          "succeed from empty file",
			configContent: []byte(`{}`),
			setupViper: func(ctrl *gomock.Controller) Viper {
				return viper.New()
			},
			want: &remoteSourceConfig{
				Provider:      defaultRemoteConfigProvider,
				Endpoint:      defaultRemoteConfigEndpoint,
				Path:          defaultRemoteConfigPath,
				SecretKeyring: defaultRemoteConfigSecretKeyring,
				Data:          "",
			},
			wantErr: false,
		},
		{
			name: "succeed from populated file",
			configContent: []byte(`
{
  "remoteConfigProvider": "external",
  "remoteConfigEndpoint": "remote:1234",
  "remoteConfigPath": "/config/path",
  "remoteConfigSecretKeyring": "super_secret"
}`),
			setupViper: func(ctrl *gomock.Controller) Viper {
				return viper.New()
			},
			want: &remoteSourceConfig{
				Provider:      "external",
				Endpoint:      "remote:1234",
				Path:          "/config/path",
				SecretKeyring: "super_secret",
				Data:          "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var configDir string
			var err error
			if tt.configContent != nil {
				configDir, err = ioutil.TempDir("", "test-parse-*")
				require.NoError(t, err, "failed creating temp config dir: %v", err)
				defer func() { _ = os.RemoveAll(configDir) }()

				tmpFilePath := filepath.Join(configDir, "config.json")
				require.NoError(t, ioutil.WriteFile(tmpFilePath, tt.configContent, 0600), "failed writing temp config file: %v", err)
			}

			v := tt.setupViper(ctrl)

			var testCfg testConfig
			got, err := loadLocalConfig(v, "test_name", configDir, "test", &testCfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadLocalConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadLocalConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validationError(t *testing.T) {
	err := validationError("provider", "prefix", "var")
	require.EqualError(t, err, "provider config provider requires PREFIX_VAR to be set")
}

// nolint:gocognit
func Test_loadRemoteConfig(t *testing.T) {
	tests := []struct {
		name              string
		setupConfigSource func() *remoteSourceConfig
		setupLocalViper   func(ctrl *gomock.Controller) Viper
		setupRemoteViper  func(ctrl *gomock.Controller) Viper
		want              *testConfig
		wantErr           bool
	}{
		{
			name: "completes configuration flow",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{}
			},
			setupLocalViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().AllKeys().Return([]string{
					keyLogLevel,
				})
				mock.EXPECT().Get(keyLogLevel).Return("DEBUG")
				return mock
			},
			setupRemoteViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyLogLevel, gomock.Any())
				mock.EXPECT().SetConfigType(defaultConfigType)
				mock.EXPECT().Unmarshal(gomock.Any())
				return mock
			},
		},
		{
			name: "fails with unmarshal error",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{}
			},
			setupLocalViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().AllKeys().Return([]string{
					keyLogLevel,
				})
				mock.EXPECT().Get(keyLogLevel).Return("DEBUG")
				return mock
			},
			setupRemoteViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyLogLevel, gomock.Any())
				mock.EXPECT().SetConfigType(defaultConfigType)
				mock.EXPECT().Unmarshal(gomock.Any()).Return(fmt.Errorf("unmarshal error"))
				return mock
			},
			wantErr: true,
		},
		{
			name: "fails with load envvar error",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "envvar",
				}
			},
			setupLocalViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().AllKeys().Return([]string{
					keyLogLevel,
				})
				mock.EXPECT().Get(keyLogLevel).Return("DEBUG")
				return mock
			},
			setupRemoteViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyLogLevel, gomock.Any())
				mock.EXPECT().SetConfigType(defaultConfigType)
				return mock
			},
			wantErr: true,
		},
		{
			name: "fails with load remote error",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "remote",
				}
			},
			setupLocalViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().AllKeys().Return([]string{
					keyLogLevel,
				})
				mock.EXPECT().Get(keyLogLevel).Return("DEBUG")
				return mock
			},
			setupRemoteViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyLogLevel, gomock.Any())
				mock.EXPECT().SetConfigType(defaultConfigType)
				return mock
			},
			wantErr: true,
		},
		{
			name: "loads a valid configuration",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{}
			},
			setupLocalViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().AllKeys().Return([]string{
					keyLogLevel,
				})
				mock.EXPECT().Get(keyLogLevel).Return("DEBUG")
				return mock
			},
			setupRemoteViper: func(ctrl *gomock.Controller) Viper {
				return viper.New()
			},
			want: &testConfig{
				BaseConfig: BaseConfig{
					Log: LogConfig{
						Level: "DEBUG",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			rsCfg := tt.setupConfigSource()

			lv := tt.setupLocalViper(ctrl)
			rv := tt.setupRemoteViper(ctrl)

			var testCfg testConfig
			if err := loadRemoteConfig(lv, rv, rsCfg, "test", &testCfg); (err != nil) != tt.wantErr {
				t.Errorf("loadRemoteConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.want != nil {
				if !reflect.DeepEqual(fmt.Sprintf("%+v", &testCfg), fmt.Sprintf("%+v", tt.want)) {
					t.Errorf("loadLocalConfig() got = %+v, want %+v", &testCfg, tt.want)
				}
			}
		})
	}
}

func Test_loadFromEnvVarSource(t *testing.T) {
	tests := []struct {
		name              string
		setupConfigSource func() *remoteSourceConfig
		setupMocks        func(mv *MockViper)
		wantErr           bool
	}{
		{
			name: "fails with missing data in configuration",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "envvar",
				}
			},
			wantErr: true,
		},
		{
			name: "fail with badly encoded envvar data",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "envvar",
					Data:     "AAA@",
				}
			},
			wantErr: true,
		},
		{
			name: "fail with envvar data read config error",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "envvar",
					Data:     "AA==",
				}
			},
			setupMocks: func(mv *MockViper) {
				mv.EXPECT().ReadConfig(gomock.Any()).Return(fmt.Errorf("read config error"))
			},
			wantErr: true,
		},
		{
			name: "succeed with envvar data",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "envvar",
					Data:     "AA==",
				}
			},
			setupMocks: func(mv *MockViper) {
				mv.EXPECT().ReadConfig(gomock.Any())
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			v := NewMockViper(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(v)
			}
			if err := loadFromEnvVarSource(v, tt.setupConfigSource(), "test"); (err != nil) != tt.wantErr {
				t.Errorf("loadFromEnvVarSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadFromRemoteSource(t *testing.T) {
	tests := []struct {
		name              string
		setupConfigSource func() *remoteSourceConfig
		setupMocks        func(mv *MockViper)
		wantErr           bool
	}{
		{
			name: "fails with missing endpoint in configuration",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider:      "remote",
					Endpoint:      "",
					Path:          "/config",
					SecretKeyring: "",
				}
			},
			wantErr: true,
		},
		{
			name: "fails with missing path in configuration",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider:      "remote",
					Endpoint:      "remote:1234",
					Path:          "",
					SecretKeyring: "",
				}
			},
			wantErr: true,
		},
		{
			name: "fails adding remote provider",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider:      "remote",
					Endpoint:      "remote:1234",
					Path:          "/config",
					SecretKeyring: "",
				}
			},
			setupMocks: func(mv *MockViper) {
				mv.EXPECT().AddRemoteProvider("remote", "remote:1234", "/config").
					Return(fmt.Errorf("provider error"))
			},
			wantErr: true,
		},
		{
			name: "fails adding secure remote provider",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider:      "remote",
					Endpoint:      "remote:1234",
					Path:          "/config",
					SecretKeyring: "keyring",
				}
			},
			setupMocks: func(mv *MockViper) {
				mv.EXPECT().AddSecureRemoteProvider("remote", "remote:1234", "/config", "keyring").
					Return(fmt.Errorf("provider error"))
			},
			wantErr: true,
		},
		{
			name: "fails reading remote provider config",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "remote",
					Endpoint: "remote:1234",
					Path:     "/config",
				}
			},
			setupMocks: func(mv *MockViper) {
				mv.EXPECT().AddRemoteProvider("remote", "remote:1234", "/config")
				mv.EXPECT().ReadRemoteConfig().Return(fmt.Errorf("read remote error"))
			},
			wantErr: true,
		},
		{
			name: "fails with invalid source config",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "remote",
					Endpoint: "",
				}
			},
			wantErr: true,
		},
		{
			name: "succeed reading remote provider config",
			setupConfigSource: func() *remoteSourceConfig {
				return &remoteSourceConfig{
					Provider: "remote",
					Endpoint: "remote:1234",
					Path:     "/config",
				}
			},
			setupMocks: func(mv *MockViper) {
				mv.EXPECT().AddRemoteProvider("remote", "remote:1234", "/config")
				mv.EXPECT().ReadRemoteConfig()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			v := NewMockViper(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(v)
			}
			if err := loadFromRemoteSource(v, tt.setupConfigSource(), "test"); (err != nil) != tt.wantErr {
				t.Errorf("loadFromRemoteSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// nolint:gocognit
func TestLoad(t *testing.T) {
	tests := []struct {
		name             string
		setupLocalViper  func(ctrl *gomock.Controller) Viper
		setupRemoteViper func(ctrl *gomock.Controller) Viper
		configContent    []byte
		envDataContent   []byte
		targetConfig     *testConfig
		wantConfig       *testConfig
		wantErr          bool
	}{
		{
			name: "success with all types and min type values",
			configContent: []byte(`
{
  "log": {
    "format": "json",
    "level": "debug",
    "network": "log_network",
    "address": "log_address"
  },
  "bool": false,
  "string": "string_value",
  "int": -12345,
  "int8": -128,
  "int16": -32768,
  "int32": -2147483648,
  "int64": -9007199254740991,
  "uint": 0,
  "uint8": 0,
  "uint16": 0,
  "uint32": 0,
  "uint64": 0,
  "float32": -1234567.89,
  "float64": -1234567890.123,
  "slice": [
    "a",
    "b"
  ],
  "mapstring": {
    "k1": "v1",
    "k2": "v2"
  },
  "nestedmap": {
    "k1": {
      "ck1": "v1",
      "ck2": "v2"
    }
  },
  "data": {
    "str": "data_string",
    "int": 12345,
    "arr": [
      1,
      2,
      3
    ]
  },
  "mapdata": {
    "d1": {
      "str": "data_string",
      "int": 56789,
      "arr": [
        4,
        5,
        6
      ]
    }
  }
}
`),
			targetConfig: &testConfig{},
			wantConfig: &testConfig{
				BaseConfig: BaseConfig{
					Log: LogConfig{
						Format:  "json",
						Level:   "debug",
						Network: "log_network",
						Address: "log_address",
					},
				},
				Boolean: false,
				String:  "string_value",
				Int:     -12345,
				Int8:    -128,
				Int16:   -32768,
				Int32:   -2147483648,
				Int64:   -9007199254740991,
				UInt:    0,
				UInt8:   0,
				UInt16:  0,
				UInt32:  0,
				UInt64:  0,
				Float32: -1234567.89,
				Float64: -1234567890.123,
				Slice:   []string{"a", "b"},
				MapString: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
				NestedMap: map[string]map[string]string{
					"k1": {
						"ck1": "v1",
						"ck2": "v2",
					},
				},
				Data: testData{
					Str: "data_string",
					Int: 12345,
					Arr: []int{1, 2, 3},
				},
				MapData: map[string]testData{
					"d1": {
						Str: "data_string",
						Int: 56789,
						Arr: []int{4, 5, 6},
					},
				},
			},
		},
		{
			name: "success with all types and max type values",
			configContent: []byte(`
{
  "log": {
    "format": "json",
    "level": "debug",
    "network": "log_network",
    "address": "log_address"
  },
  "bool": true,
  "string": "string_value",
  "int": 12345,
  "int8": 127,
  "int16": 32767,
  "int32": 2147483647,
  "int64": 9007199254740991,
  "uint": 12345,
  "uint8": 255,
  "uint16": 65535,
  "uint32": 4294967295,
  "uint64": 9007199254740991,
  "float32": 1234567.89,
  "float64": 1234567890.123,
  "slice": [
    "a",
    "b"
  ],
  "mapstring": {
    "k1": "v1",
    "k2": "v2"
  },
  "nestedmap": {
    "k1": {
      "ck1": "v1",
      "ck2": "v2"
    }
  },
  "data": {
    "str": "data_string",
    "int": 12345,
    "arr": [
      1,
      2,
      3
    ]
  },
  "mapdata": {
    "d1": {
      "str": "data_string",
      "int": 56789,
      "arr": [
        4,
        5,
        6
      ]
    }
  }
}
`),
			targetConfig: &testConfig{},
			wantConfig: &testConfig{
				BaseConfig: BaseConfig{
					Log: LogConfig{
						Format:  "json",
						Level:   "debug",
						Network: "log_network",
						Address: "log_address",
					},
				},
				Boolean: true,
				String:  "string_value",
				Int:     12345,
				Int8:    127,
				Int16:   32767,
				Int32:   2147483647,
				Int64:   9007199254740991,
				UInt:    12345,
				UInt8:   255,
				UInt16:  65535,
				UInt32:  4294967295,
				UInt64:  9007199254740991,
				Float32: 1234567.89,
				Float64: 1234567890.123,
				Slice:   []string{"a", "b"},
				MapString: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
				NestedMap: map[string]map[string]string{
					"k1": {
						"ck1": "v1",
						"ck2": "v2",
					},
				},
				Data: testData{
					Str: "data_string",
					Int: 12345,
					Arr: []int{1, 2, 3},
				},
				MapData: map[string]testData{
					"d1": {
						Str: "data_string",
						Int: 56789,
						Arr: []int{4, 5, 6},
					},
				},
			},
		},
		{
			name: "fails loading local config",
			setupLocalViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(keyRemoteConfigProvider, defaultRemoteConfigProvider)
				mock.EXPECT().SetDefault(keyRemoteConfigEndpoint, defaultRemoteConfigEndpoint)
				mock.EXPECT().SetDefault(keyRemoteConfigPath, defaultRemoteConfigPath)
				mock.EXPECT().SetDefault(keyRemoteConfigSecretKeyring, defaultRemoteConfigSecretKeyring)
				mock.EXPECT().SetDefault(keyLogFormat, defaultLogFormat)
				mock.EXPECT().SetDefault(keyLogLevel, defaultLogLevel)
				mock.EXPECT().SetDefault(keyLogAddress, defaultLogAddress)
				mock.EXPECT().SetDefault(keyLogNetwork, defaultLogNetwork)
				mock.EXPECT().SetDefault("drink", "c0ffee")
				mock.EXPECT().SetConfigName(defaultConfigName)
				mock.EXPECT().SetConfigType(defaultConfigType)
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().AutomaticEnv()
				mock.EXPECT().SetEnvPrefix("test")
				mock.EXPECT().BindEnv(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig().Return(fmt.Errorf("read config error"))
				return mock
			},
			targetConfig: &testConfig{},
			wantErr:      true,
		},
		{
			name: "fails loading remote config",
			setupRemoteViper: func(ctrl *gomock.Controller) Viper {
				mock := NewMockViper(ctrl)
				mock.EXPECT().SetDefault(gomock.Any(), gomock.Any()).AnyTimes()
				mock.EXPECT().SetConfigType(defaultConfigType)
				mock.EXPECT().Unmarshal(gomock.Any()).Return(fmt.Errorf("unmarshal error"))
				return mock
			},
			configContent: []byte(`
{
  "log": {
    "format": "json",
    "level": "debug",
    "network": "log_network",
    "address": "log_address"
  },
  "data": {
    "str": "data_string",
    "int": 12345,
    "arr": [
      1,
      2,
      3
    ]
  }
}
`),
			targetConfig: &testConfig{},
			wantErr:      true,
		},
		{
			name: "fails validating configuration",
			configContent: []byte(`
{
  "log": {
    "format": "json",
    "level": "debug",
    "network": "log_network",
    "address": "log_address"
  },
  "data": {
    "str": "data_string",
    "int": 12345,
    "arr": [
      1,
      2,
      3
    ]
  }
}
`),
			targetConfig: &testConfig{validateErr: fmt.Errorf("validate error")},
			wantErr:      true,
		},
		{
			name: "success load local config w/o remote override",
			configContent: []byte(`
{
  "log": {
	"format": "json",
	"level": "debug",
	"network": "log_network",
	"address": "log_address"
  },
  "data": {
	"str": "data_string",
	"int": 12345,
	"arr": [
	  1,
	  2,
	  3
	]
  }
}
`),
			targetConfig: &testConfig{},
			wantConfig: &testConfig{
				BaseConfig: BaseConfig{
					Log: LogConfig{
						Format:  "json",
						Level:   "debug",
						Network: "log_network",
						Address: "log_address",
					},
				},
				Data: testData{
					Str: "data_string",
					Int: 12345,
					Arr: []int{1, 2, 3},
				},
			},
		},
		{
			name: "success load local config w/ envar override",
			configContent: []byte(`
{
  "remoteConfigProvider": "envvar",
  "log": {
	"format": "console",
	"level": "info"
  },
  "data": {
	"str": "data_string",
	"int": 12345,
	"arr": [
	  1,
	  2,
	  3
	]
  }
}
`),
			envDataContent: []byte(`
{
  "log": {
	"format": "json",
	"level": "debug",
	"network": "log_network",
	"address": "log_address"
  },
  "data": {
	"str": "data_string",
	"int": 56,
	"arr": [
	  4,
	  5,
	  6
	]
  }
}
`),
			targetConfig: &testConfig{},
			wantConfig: &testConfig{
				BaseConfig: BaseConfig{
					Log: LogConfig{
						Format:  "json",
						Level:   "debug",
						Network: "log_network",
						Address: "log_address",
					},
				},
				Data: testData{
					Str: "data_string",
					Int: 56,
					Arr: []int{4, 5, 6},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var tmpConfigDir string
			var err error
			if tt.configContent != nil {
				tmpConfigDir, err = ioutil.TempDir("", "test-parse-*")
				require.NoError(t, err, "failed creating temp config dir: %v", err)
				defer func() { _ = os.RemoveAll(tmpConfigDir) }()

				tmpFilePath := filepath.Join(tmpConfigDir, "config.json")
				require.NoError(t, ioutil.WriteFile(tmpFilePath, tt.configContent, 0600), "failed writing temp config file: %v", err)
			}

			if tt.setupLocalViper != nil {
				localViper = tt.setupLocalViper(ctrl)
			} else {
				localViper = viper.New()
			}

			if tt.setupRemoteViper != nil {
				remoteViper = tt.setupRemoteViper(ctrl)
			} else {
				remoteViper = viper.New()
			}

			if tt.envDataContent != nil {
				envKey := strings.ToUpper(fmt.Sprintf("%s_REMOTECONFIGDATA", "test"))
				_ = os.Setenv(envKey, base64.StdEncoding.EncodeToString(tt.envDataContent))
				defer func() { _ = os.Unsetenv(envKey) }()
			}

			if err := Load("cmd", tmpConfigDir, "test", tt.targetConfig); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantConfig != nil {
				cfgStr := func() string {
					data, _ := json.Marshal(tt.targetConfig)
					return string(data)
				}()
				wantCfgStr := func() string {
					data, _ := json.Marshal(tt.wantConfig)
					return string(data)
				}()
				if cfgStr != wantCfgStr {
					t.Errorf("loadRemoteSourceConfig() got = %s, want = %s", cfgStr, wantCfgStr)
				}
			}
		})
	}
}

func TestReset(t *testing.T) {
	oldLocalViper := localViper
	oldRemoteViper := remoteViper

	Reset()

	require.NotEqual(t, oldLocalViper, localViper)
	require.NotEqual(t, oldRemoteViper, remoteViper)
}
