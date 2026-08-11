package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/UKatkar/order-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func (r *RedisRepo) InsertOrder(ctx context.Context, order model.Order) error {

	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	tx := r.Client.TxPipeline()

	key := orderIdKey(order.OrderID)

	err = tx.SetNX(ctx, key, string(data), 0).Err()
	if err != nil {
		tx.Discard()
		return fmt.Errorf("failed to insert order into redis: %w", err)
	}

	if err := tx.SAdd(ctx, "orders", key).Err(); err != nil {
		tx.Discard()
		return fmt.Errorf("failed to add order key to orders set: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("order not exist")

func (r *RedisRepo) FindOrderById(ctx context.Context, orderId uint64) (*model.Order, error) {
	key := orderIdKey(orderId)

	data, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return &model.Order{}, ErrNotExist
	} else if err != nil {
		return nil, fmt.Errorf("failed to get order from redis: %w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(data), &order)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

func (r *RedisRepo) DeleteOrderById(ctx context.Context, orderId uint64) error {
	key := orderIdKey(orderId)

	tx := r.Client.TxPipeline()

	err := tx.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		tx.Discard()
		return ErrNotExist
	}

	if err != nil {
		tx.Discard()
		return fmt.Errorf("failed to delete order from redis: %w", err)
	}

	if err := tx.S(ctx, "orders", key).Err(); err != nil {
		tx.Discard()
		return fmt.Errorf("failed to remove order key from orders set: %w", err)
	}

	return nil
}

func (r *RedisRepo) UpdateOrder(ctx context.Context, order model.Order) error {

	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = r.Client.SetXX(ctx, string(data), 0).Err()

	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	}

	if err != nil {
		return fmt.Errorf("failed to update order in redis: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAllOrders(ctx context.Context, page FindAllPage) (FindResult, error) {

	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to scan orders: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders from redis: %w", err)
	}

	orders := make([]model.Order, 0, len(xs))

	for _, x := range xs {
		if x == nil {
			continue
		}

		var order model.Order
		err = json.Unmarshal([]byte(x.(string)), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to unmarshal order: %w", err)
		}
		orders = append(orders, order)
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}

func orderIdKey(orderId uint64) string {
	return fmt.Sprintf("order:%d", orderId)
}
