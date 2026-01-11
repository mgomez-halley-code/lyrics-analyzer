package service

import (
	"context"
	"testing"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/cache"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of LyricsProvider
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	args := m.Called(ctx, track, artist)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.LyricsSourceData), args.Error(1)
}

func TestCachingProvider_CacheHit(t *testing.T) {
	mockProvider := new(MockProvider)
	cache := cache.NewNoOpCache()
	cachingProvider := NewCachingProvider(mockProvider, cache, time.Hour)

	testData := &model.LyricsSourceData{
		TrackID:    123,
		TrackName:  "Test Song",
		ArtistName: "Test Artist",
	}

	// First call - should hit the provider
	mockProvider.On("GetLyrics", mock.Anything, "Test Song", "Test Artist").
		Return(testData, nil).Once()

	result, err := cachingProvider.GetLyrics(context.Background(), "Test Song", "Test Artist")
	assert.NoError(t, err)
	assert.Equal(t, testData, result)

	// Verify mock was called
	mockProvider.AssertExpectations(t)
}

func TestCachingProvider_GenerateCacheKey(t *testing.T) {
	cache := cache.NewNoOpCache()
	cp := NewCachingProvider(nil, cache, time.Hour)

	// Test key generation
	key1 := cp.generateCacheKey("Imagine", "John Lennon")
	key2 := cp.generateCacheKey("Imagine", "John Lennon")
	key3 := cp.generateCacheKey("imagine", "john lennon")     // Case insensitive
	key4 := cp.generateCacheKey(" Imagine ", " John Lennon ") // Trim spaces

	// Same song should generate same key
	assert.Equal(t, key1, key2)
	assert.Equal(t, key1, key3)
	assert.Equal(t, key1, key4)

	// Different song should generate different key
	key5 := cp.generateCacheKey("Yesterday", "The Beatles")
	assert.NotEqual(t, key1, key5)

	// Keys should start with "lyrics:"
	assert.Contains(t, key1, "lyrics:")
}
