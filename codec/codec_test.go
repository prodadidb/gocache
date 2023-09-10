package codec_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prodadidb/gocache/codec"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock_store_interface_test.go -package=codec_test -source=../store/interface.go

func TestNew(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	// When
	c := codec.New(s)

	// Then
	assert.IsType(t, new(codec.Codec), c)
}

func TestGetWhenHit(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)

	c := codec.New(s)

	// When
	value, err := c.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)

	assert.Equal(t, 1, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestGetWithTTLWhenHit(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue, 1*time.Second, nil)

	c := codec.New(s)

	// When
	value, ttl, err := c.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, 1*time.Second, ttl)

	assert.Equal(t, 1, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestGetWithTTLWhenMiss(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to find in store")

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second, expectedErr)

	c := codec.New(s)

	// When
	value, ttl, err := c.GetWithTTL(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 1, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestGetWhenMiss(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to find in store")

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Get(ctx, "my-key").Return(nil, expectedErr)

	c := codec.New(s)

	// When
	value, err := c.Get(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 1, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestSetWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Set(ctx, "my-key", cacheValue, store.OptionsMatcher{
		Expiration: 5 * time.Second,
	}).Return(nil)

	c := codec.New(mockedStore)

	// When
	err := c.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 1, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	expectedErr := errors.New("unable to set value in store")

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Set(ctx, "my-key", cacheValue, store.OptionsMatcher{
		Expiration: 5 * time.Second,
	}).Return(expectedErr)

	c := codec.New(mockedStore)

	// When
	err := c.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 1, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestDeleteWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Delete(ctx, "my-key").Return(nil)

	c := codec.New(s)

	// When
	err := c.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 1, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	c := codec.New(s)

	// When
	err := c.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 1, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestInvalidateWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{
		Tags: []string{"tag1"},
	}).Return(nil)

	c := codec.New(mockedStore)

	// When
	err := c.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 1, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{
		Tags: []string{"tag1"},
	}).Return(expectedErr)

	c := codec.New(mockedStore)

	// When
	err := c.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 1, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestClearWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Clear(ctx).Return(nil)

	c := codec.New(s)

	// When
	err := c.Clear(ctx)

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 1, c.GetStats().ClearSuccess)
	assert.Equal(t, 0, c.GetStats().ClearError)
}

func TestClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when clearing cache")

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Clear(ctx).Return(expectedErr)

	c := codec.New(s)

	// When
	err := c.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, c.GetStats().Hits)
	assert.Equal(t, 0, c.GetStats().Miss)
	assert.Equal(t, 0, c.GetStats().SetSuccess)
	assert.Equal(t, 0, c.GetStats().SetError)
	assert.Equal(t, 0, c.GetStats().DeleteSuccess)
	assert.Equal(t, 0, c.GetStats().DeleteError)
	assert.Equal(t, 0, c.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, c.GetStats().InvalidateError)
	assert.Equal(t, 0, c.GetStats().ClearSuccess)
	assert.Equal(t, 1, c.GetStats().ClearError)
}

func TestGetStore(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	c := codec.New(s)

	// When - Then
	assert.Equal(t, s, c.GetStore())
}

func TestGetStats(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	c := codec.New(s)

	// When - Then
	expectedStats := &codec.Stats{}
	assert.Equal(t, expectedStats, c.GetStats())
}
