package repository

import (
	"context"
	"time"

	"github.com/vogiaan1904/payment-svc/internal/models"
	"github.com/vogiaan1904/payment-svc/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const paymentCollection = "payments"

type implPaymentRepository struct {
	l  log.Logger
	db *mongo.Database
}

func NewPaymentRepository(l log.Logger, db *mongo.Database) PaymentRepository {
	return &implPaymentRepository{
		l:  l,
		db: db,
	}
}

func (r *implPaymentRepository) collection() *mongo.Collection {
	return r.db.Collection(paymentCollection)
}

func (r *implPaymentRepository) CreatePayment(ctx context.Context, opt CreatePaymentOptions) (models.Payment, error) {
	r.l.Debugf(ctx, "CreatePayment: %+v", opt)

	now := time.Now()
	p := models.Payment{
		ID:               primitive.NewObjectID(),
		OrderID:          opt.OrderID,
		UserID:           opt.UserID,
		Amount:           opt.Amount,
		Currency:         opt.Currency,
		Status:           opt.Status,
		Method:           opt.Method,
		GatewayReference: opt.GatewayReference,
		Description:      opt.Description,
		Metadata:         opt.Metadata,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	_, err := r.collection().InsertOne(ctx, p)
	if err != nil {
		r.l.Errorf(ctx, "Failed to insert payment: %v", err)
		return models.Payment{}, err
	}

	return p, nil
}

func (r *implPaymentRepository) FindOnePayment(ctx context.Context, opt FindOnePaymentOptions) (models.Payment, error) {
	r.l.Debugf(ctx, "FindOnePayment: %+v", opt)

	filter := bson.M{}
	if opt.ID != "" {
		id, err := primitive.ObjectIDFromHex(opt.ID)
		if err != nil {
			r.l.Errorf(ctx, "Invalid ID: %v", err)
			return models.Payment{}, err
		}
		filter["_id"] = id
	}

	if opt.OrderID != "" {
		filter["order_id"] = opt.OrderID
	}

	if opt.UserID != "" {
		filter["user_id"] = opt.UserID
	}

	if !opt.FindFilter.IncludeDeleted {
		filter["deleted_at"] = nil
	}

	var payment models.Payment
	err := r.collection().FindOne(ctx, filter).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.l.Warnf(ctx, "Payment not found: %v", err)
		} else {
			r.l.Errorf(ctx, "Failed to find payment: %v", err)
		}
		return models.Payment{}, err
	}

	return payment, nil
}

func (r *implPaymentRepository) UpdatePaymentStatus(ctx context.Context, opt UpdatePaymentStatusOptions) error {
	r.l.Debugf(ctx, "UpdatePaymentStatus: %+v", opt)

	id, err := primitive.ObjectIDFromHex(opt.ID)
	if err != nil {
		r.l.Errorf(ctx, "Invalid ID: %v", err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":            opt.Status,
			"gateway_reference": opt.GatewayReference,
			"updated_at":        time.Now(),
		},
	}

	_, err = r.collection().UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.l.Errorf(ctx, "Failed to update payment status: %v", err)
		return err
	}

	return nil
}

func (r *implPaymentRepository) CancelPayment(ctx context.Context, opt CancelPaymentOptions) error {
	r.l.Debugf(ctx, "CancelPayment: %+v", opt)

	id, err := primitive.ObjectIDFromHex(opt.ID)
	if err != nil {
		r.l.Errorf(ctx, "Invalid ID: %v", err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     models.PaymentStatusCancelled,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection().UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.l.Errorf(ctx, "Failed to cancel payment: %v", err)
		return err
	}

	return nil
}
