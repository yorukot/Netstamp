package postgres

import (
	"context"
	"time"
)

type SeriesScope struct {
	ProbeStorageID int64
	CheckStorageID int64
	StartedAtFrom  time.Time
	StartedAtTo    time.Time
	MaxDataPoints  int32
}

type RawSeriesParams interface {
	~struct {
		ProbeStorageID int64     `json:"probe_storage_id"`
		CheckStorageID int64     `json:"check_storage_id"`
		StartedAtFrom  time.Time `json:"started_at_from"`
		StartedAtTo    time.Time `json:"started_at_to"`
	}
}

type BucketSeriesParams interface {
	~struct {
		StartedAtFrom  time.Time `json:"started_at_from"`
		ProbeStorageID int64     `json:"probe_storage_id"`
		CheckStorageID int64     `json:"check_storage_id"`
		StartedAtTo    time.Time `json:"started_at_to"`
		MaxDataPoints  float64   `json:"max_data_points"`
	}
}

type RollupSeriesParams interface {
	~struct {
		StartedAtTo    time.Time `json:"started_at_to"`
		StartedAtFrom  time.Time `json:"started_at_from"`
		MaxDataPoints  float64   `json:"max_data_points"`
		ProbeStorageID int64     `json:"probe_storage_id"`
		CheckStorageID int64     `json:"check_storage_id"`
	}
}

func NewRawSeriesQuery[Params RawSeriesParams, Row, Point any](
	query func(context.Context, Params) ([]Row, error),
	mapRows func([]Row) []Point,
) func(context.Context, SeriesScope) ([]Point, error) {
	return func(ctx context.Context, scope SeriesScope) ([]Point, error) {
		rows, err := query(ctx, Params{
			ProbeStorageID: scope.ProbeStorageID,
			CheckStorageID: scope.CheckStorageID,
			StartedAtFrom:  scope.StartedAtFrom,
			StartedAtTo:    scope.StartedAtTo,
		})
		return mapRows(rows), err
	}
}

func NewBucketSeriesQuery[Params BucketSeriesParams, Row, Point any](
	query func(context.Context, Params) ([]Row, error),
	mapRows func([]Row) []Point,
) func(context.Context, SeriesScope) ([]Point, error) {
	return func(ctx context.Context, scope SeriesScope) ([]Point, error) {
		rows, err := query(ctx, Params{
			StartedAtFrom:  scope.StartedAtFrom,
			ProbeStorageID: scope.ProbeStorageID,
			CheckStorageID: scope.CheckStorageID,
			StartedAtTo:    scope.StartedAtTo,
			MaxDataPoints:  float64(scope.MaxDataPoints),
		})
		return mapRows(rows), err
	}
}

func NewRollupSeriesQuery[Params RollupSeriesParams, Row, Point any](
	query func(context.Context, Params) ([]Row, error),
	mapRows func([]Row) []Point,
) func(context.Context, SeriesScope) ([]Point, error) {
	return func(ctx context.Context, scope SeriesScope) ([]Point, error) {
		rows, err := query(ctx, Params{
			StartedAtTo:    scope.StartedAtTo,
			StartedAtFrom:  scope.StartedAtFrom,
			MaxDataPoints:  float64(scope.MaxDataPoints),
			ProbeStorageID: scope.ProbeStorageID,
			CheckStorageID: scope.CheckStorageID,
		})
		return mapRows(rows), err
	}
}
