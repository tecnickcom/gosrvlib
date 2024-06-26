package passwordhash

import (
	"errors"
	"testing"
	"testing/iotest"

	"github.com/Vonage/gosrvlib/pkg/random"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestNew(t *testing.T) {
	t.Parallel()

	p := New()

	require.NotEmpty(t, p.Algo)
	require.NotZero(t, p.Version)
	require.NotZero(t, p.Threads)

	opts := []Option{
		WithKeyLen(31),
		WithSaltLen(17),
		WithTime(3),
		WithMemory(65_537),
		WithThreads(5),
		WithMinPasswordLength(16),
		WithMaxPasswordLength(128),
	}

	p = New(opts...)

	require.Equal(t, DefaultAlgo, p.Algo)
	require.Equal(t, uint8(argon2.Version), p.Version)
	require.Equal(t, uint32(31), p.KeyLen)
	require.Equal(t, uint32(17), p.SaltLen)
	require.Equal(t, uint32(3), p.Time)
	require.Equal(t, uint32(0xfff0), p.Memory)
	require.Equal(t, uint8(5), p.Threads)
	require.Equal(t, uint32(16), p.minPLen)
	require.Equal(t, uint32(128), p.maxPLen)
}

func Test_passwordHashData(t *testing.T) {
	t.Parallel()

	p := New()

	hash, err := p.passwordHashData("test-password")

	require.NoError(t, err)
	require.NotEmpty(t, hash)

	shortPassword := string(make([]byte, p.minPLen-1))

	hash, err = p.passwordHashData(shortPassword)

	require.Error(t, err)
	require.Empty(t, hash)

	longPassword := string(make([]byte, p.maxPLen+1))

	hash, err = p.passwordHashData(longPassword)

	require.Error(t, err)
	require.Empty(t, hash)

	p.rnd = random.New(iotest.ErrReader(errors.New("test-rand-reader-error")))

	hash, err = p.passwordHashData("test")

	require.Error(t, err)
	require.Empty(t, hash)
}

func Test_passwordHashData_passwordVerifyData(t *testing.T) {
	t.Parallel()

	p := New()

	secret := "test-secret-string"
	data, err := p.passwordHashData(secret)

	require.NoError(t, err)
	require.NotEmpty(t, data)

	ok, err := p.passwordVerifyData(secret, data)

	require.NoError(t, err)
	require.True(t, ok)

	ok, err = p.passwordVerifyData("test-wrong-secret", data)

	require.NoError(t, err)
	require.False(t, ok)

	p.Algo = "wrong-algo"

	ok, err = p.passwordVerifyData(secret, data)

	require.Error(t, err)
	require.False(t, ok)

	p.Algo = DefaultAlgo
	p.Version = 0

	ok, err = p.passwordVerifyData(secret, data)

	require.Error(t, err)
	require.False(t, ok)
}

func TestPasswordHash(t *testing.T) {
	t.Parallel()

	p := New()

	hash, err := p.PasswordHash("TestPasswordString")

	require.NoError(t, err)
	require.NotEmpty(t, hash)

	p.rnd = random.New(iotest.ErrReader(errors.New("test-rand-reader-error")))

	_, err = p.PasswordHash("test")

	require.Error(t, err)
}

func TestPasswordVerify(t *testing.T) {
	t.Parallel()

	p := New()

	hash := "eyJQIjp7IkEiOiJhcmdvbjJpZCIsIlYiOjE5LCJLIjozMiwiUyI6MTYsIlQiOjEsIk0iOjY1NTM2LCJQIjoxNn0sIlMiOiI1d25uaXRVaGV6cjFnbkdoeU1FVTdBPT0iLCJLIjoiQmNiUlRVNFNDcmQxNGJWUzRzcVBGYndvbnYreWlvZ09ueGJWMXBRTGRWMD0ifQo="

	ok, err := p.PasswordVerify("test", hash)

	require.NoError(t, err)
	require.True(t, ok)

	ok, err = p.PasswordVerify("secret", "wrong-hash")

	require.Error(t, err)
	require.False(t, ok)
}

func Test_PasswordHash_PasswordVerify(t *testing.T) {
	t.Parallel()

	secret := "Test-Password-01234"

	p := New()

	hash, err := p.PasswordHash(secret)

	require.NoError(t, err)
	require.NotEmpty(t, hash)

	ok, err := p.PasswordVerify(secret, hash)

	require.NoError(t, err)
	require.True(t, ok)
}

func Test_EncryptPasswordHash(t *testing.T) {
	t.Parallel()

	p := New()

	key := []byte("0123456789012345")
	secret := "test-secret"

	hash, err := p.EncryptPasswordHash(key, secret)

	require.NoError(t, err)
	require.NotEmpty(t, hash)

	p.rnd = random.New(iotest.ErrReader(errors.New("test-rand-reader-error")))

	hash, err = p.EncryptPasswordHash(key, secret)

	require.Error(t, err)
	require.Empty(t, hash)
}

func Test_EncryptPasswordVerify(t *testing.T) {
	t.Parallel()

	p := New()

	key := []byte("0123456789012345")
	secret := "test-secret"

	hash, err := p.EncryptPasswordHash(key, secret)

	require.NoError(t, err)
	require.NotEmpty(t, hash)

	ok, err := p.EncryptPasswordVerify(key, secret, hash)

	require.NoError(t, err)
	require.True(t, ok)

	ok, err = p.EncryptPasswordVerify(key, "wrong-secret", hash)

	require.NoError(t, err)
	require.False(t, ok)

	ok, err = p.EncryptPasswordVerify([]byte("abcdefghijklmnop"), secret, hash)

	require.Error(t, err)
	require.False(t, ok)
}
