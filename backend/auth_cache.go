package main

import (
	"sync"
	"time"
)

type CachedAuth struct {
	User      User
	ExpiresAt time.Time
}

type AuthCache struct {
	cache sync.Map
	ttl   time.Duration
}

func NewAuthCache(ttl time.Duration) *AuthCache {
	return &AuthCache{
		ttl: ttl,
	}
}

func (c *AuthCache) Get(sessionCookie string) (User, bool) {
	if val, ok := c.cache.Load(sessionCookie); ok {
		cached := val.(CachedAuth)
		if time.Now().Before(cached.ExpiresAt) {
			return cached.User, true
		}
		c.cache.Delete(sessionCookie)
	}
	return User{}, false
}

func (c *AuthCache) Set(sessionCookie string, user User) {
	c.cache.Store(sessionCookie, CachedAuth{
		User:      user,
		ExpiresAt: time.Now().Add(c.ttl),
	})
}
