// Package cache provides a Redis client and a set of methods for interacting with Redis.
//
// The cache package abstracts away the details of interacting with Redis, allowing the rest of
// the application to use a simpler, higher-level API for accessing the data stored in cache. By centralizing Redis access
// in the cache package, it helps to ensure consistency and maintainability of the cache access layer.
//
// Caching frequently accessed data can improve application performance and reduce database load. Using the cache
// package simplifies the process of implementing caching in the application.
package cache
