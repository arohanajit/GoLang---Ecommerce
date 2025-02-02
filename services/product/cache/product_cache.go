// services/product/cache/product_cache.go
package cache

import (
	"context"
	"fmt"
	"time"

	"e-commerce-platform/pkg/redis"
)

type ProductCache struct {
	redisClient *redis.Client
}

func NewProductCache(redisClient *redis.Client) *ProductCache {
	return &ProductCache{redisClient: redisClient}
}

func (pc *ProductCache) GetProduct(ctx context.Context, productID string) (*Product, error) {
	var product Product
	err := pc.redisClient.Get(ctx, fmt.Sprintf("product:%s", productID), &product)
	return &product, err
}

func (pc *ProductCache) SetProduct(ctx context.Context, product *Product) error {
	return pc.redisClient.Set(ctx, fmt.Sprintf("product:%s", product.ID), product, 15*time.Minute)
}

func (pc *ProductCache) InvalidateProduct(ctx context.Context, productID string) error {
	return pc.redisClient.Delete(ctx, fmt.Sprintf("product:%s", productID))
}

func (pc *ProductCache) GetProductsList(ctx context.Context, page, limit int) ([]*Product, error) {
	var products []*Product
	key := fmt.Sprintf("products:list:%d:%d", page, limit)
	err := pc.redisClient.Get(ctx, key, &products)
	return products, err
}

func (pc *ProductCache) SetProductsList(ctx context.Context, products []*Product, page, limit int) error {
	key := fmt.Sprintf("products:list:%d:%d", page, limit)
	return pc.redisClient.Set(ctx, key, products, 5*time.Minute)
}
