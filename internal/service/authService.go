package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
)

// ==================== TOKEN MANAGEMENT ====================

func UpsertTokenData(userId string, dbData bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbData["updated_at"] = time.Now()

	opts := options.UpdateOne().SetUpsert(true)
	result, err := database.DbAuth.Collection("access_tokens").UpdateOne(
		ctx,
		bson.M{"user_id": userId},
		bson.M{"$set": dbData},
		opts,
	)

	if err != nil {
		middleware.LogError(err, "MongoDB Failed to Save Token Data")
		return err
	}

	if result.UpsertedCount > 0 {
		fmt.Printf("✅ Token inserted for user_id: %s\n", userId)
	} else if result.ModifiedCount > 0 {
		fmt.Printf("✅ Token updated for user_id: %s\n", userId)
	}

	return nil
}

func DeleteTokenData(userId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := database.DbAuth.Collection("access_tokens").DeleteOne(
		ctx,
		bson.M{"user_id": userId},
	)

	if err != nil {
		middleware.LogError(err, "MongoDB Failed to Delete Token Data")
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("token not found")
	}

	fmt.Printf("✅ Token deleted for user_id: %s\n", userId)
	return nil
}

func GetTokenData(whereParam bson.M) (*model.TokenData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var storedToken model.TokenData
	err := database.DbAuth.Collection("access_tokens").FindOne(ctx, whereParam).Decode(&storedToken)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("token not found")
		}
		middleware.LogError(err, "MongoDB Failed to Retrieve Token Data")
		return nil, err
	}

	return &storedToken, nil
}

func InvalidateToken(userId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_valid_token": "n",
			"updated_at":     time.Now(),
		},
	}

	result, err := database.DbAuth.Collection("access_tokens").UpdateOne(
		ctx,
		bson.M{"user_id": userId},
		update,
	)

	if err != nil {
		middleware.LogError(err, "MongoDB Failed to Invalidate Token")
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("token not found")
	}

	fmt.Printf("🔒 Token invalidated for user_id: %s\n", userId)
	return nil
}

func InvalidateAllUserTokens(userId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_valid_token": "n",
			"updated_at":     time.Now(),
		},
	}

	_, err := database.DbAuth.Collection("access_tokens").UpdateMany(
		ctx,
		bson.M{"user_id": userId},
		update,
	)

	if err != nil {
		middleware.LogError(err, "MongoDB Failed to Invalidate All Tokens")
		return err
	}

	fmt.Printf("🔒 All tokens invalidated for user_id: %s\n", userId)
	return nil
}

func RefreshAccessToken(userId string, newAccessToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"access_token": newAccessToken,
			"updated_at":   time.Now(),
		},
	}

	result, err := database.DbAuth.Collection("access_tokens").UpdateOne(
		ctx,
		bson.M{"user_id": userId, "is_valid_token": "y"},
		update,
	)

	if err != nil {
		middleware.LogError(err, "MongoDB Failed to Refresh Access Token")
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("valid token not found for user")
	}

	fmt.Printf("✅ Access token refreshed for user_id: %s\n", userId)
	return nil
}

func ValidateRefreshToken(userId string, refreshToken string) (*model.TokenData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var tokenData model.TokenData
	err := database.DbAuth.Collection("access_tokens").FindOne(
		ctx,
		bson.M{
			"user_id":        userId,
			"refresh_token":  refreshToken,
			"is_valid_token": "y",
		},
	).Decode(&tokenData)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid or expired refresh token")
		}
		middleware.LogError(err, "MongoDB Failed to Validate Refresh Token")
		return nil, err
	}

	return &tokenData, nil
}

func CleanupExpiredTokens(daysOld int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoffDate := time.Now().AddDate(0, 0, -daysOld)

	result, err := database.DbAuth.Collection("access_tokens").DeleteMany(
		ctx,
		bson.M{
			"updated_at": bson.M{"$lt": cutoffDate},
		},
	)

	if err != nil {
		middleware.LogError(err, "MongoDB Failed to Cleanup Expired Tokens")
		return err
	}

	fmt.Printf("✅ Cleaned up %d expired tokens\n", result.DeletedCount)
	return nil
}

func GetAllActiveTokens(userId string) ([]model.TokenData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.DbAuth.Collection("access_tokens").Find(
		ctx,
		bson.M{"user_id": userId, "is_valid_token": "y"},
	)

	if err != nil {
		middleware.LogError(err, "MongoDB Failed to Get Active Tokens")
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []model.TokenData
	if err = cursor.All(ctx, &tokens); err != nil {
		middleware.LogError(err, "MongoDB Failed to Decode Tokens")
		return nil, err
	}

	return tokens, nil
}

// ==================== REFERRAL TRACKING IN AUTH ====================

// TrackReferralLogin mencatat login dari user yang direferral
func TrackReferralLogin(userId uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := GetOneUser(fmt.Sprintf("%d", userId))
	if user == nil {
		return errors.New("user not found")
	}
	loginRecord := bson.M{
		"user_id":    userId,
		"login_at":   time.Now(),
		"created_at": time.Now(),
	}

	_, err := database.DbAuth.Collection("login_history").InsertOne(ctx, loginRecord)
	if err != nil {
		middleware.LogError(err, "Failed to track referral login")
		return err
	}

	fmt.Printf("📊 Login tracked for user_id: %d\n", userId)
	return nil
}

// GetReferralLoginStats mendapatkan statistik login dari referral users
func GetReferralLoginStats(referralCode string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Dapatkan users yang direferral oleh kode ini
	referredUsers, err := GetUserReferrals(referralCode)
	if err != nil {
		return nil, err
	}

	userIds := make([]uint, len(referredUsers))
	for i, user := range referredUsers {
		userIds[i] = user.ID
	}

	// Hitung total login dari referred users
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": bson.M{"$in": userIds}}}},
		{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"total_logins": bson.M{"$sum": 1},
			"last_login":   bson.M{"$max": "$login_at"},
		}}},
	}

	cursor, err := database.DbAuth.Collection("login_history").Aggregate(ctx, pipeline)
	if err != nil {
		middleware.LogError(err, "Failed to get referral login stats")
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		middleware.LogError(err, "Failed to decode login stats")
		return nil, err
	}

	stats := map[string]interface{}{
		"referral_code":      referralCode,
		"total_referred":     len(referredUsers),
		"total_logins":       0,
		"last_login":         nil,
	}

	if len(results) > 0 {
		stats["total_logins"] = results[0]["total_logins"]
		stats["last_login"] = results[0]["last_login"]
	}

	return stats, nil
}

// GetUserLoginHistory mendapatkan riwayat login user
func GetUserLoginHistory(userId uint, limit int) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "login_at", Value: -1}}).SetLimit(int64(limit))
	
	cursor, err := database.DbAuth.Collection("login_history").Find(
		ctx,
		bson.M{"user_id": userId},
		opts,
	)

	if err != nil {
		middleware.LogError(err, "Failed to get user login history")
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []bson.M
	if err = cursor.All(ctx, &history); err != nil {
		middleware.LogError(err, "Failed to decode login history")
		return nil, err
	}

	return history, nil
}

// GetReferralEngagementMetrics mendapatkan metrik engagement dari referral
func GetReferralEngagementMetrics(referralCode string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get referred users
	referredUsers, err := GetUserReferrals(referralCode)
	if err != nil {
		return nil, err
	}

	if len(referredUsers) == 0 {
		return map[string]interface{}{
			"referral_code":    referralCode,
			"total_referred":   0,
			"active_users":     0,
			"engagement_rate":  0,
		}, nil
	}

	userIds := make([]uint, len(referredUsers))
	activeCount := 0
	
	for i, user := range referredUsers {
		userIds[i] = user.ID
		if user.IsAktif == "active" {
			activeCount++
		}
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id":  bson.M{"$in": userIds},
			"login_at": bson.M{"$gte": thirtyDaysAgo},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$user_id",
		}}},
		{{Key: "$count", Value: "active_last_30_days"}},
	}

	cursor, err := database.DbAuth.Collection("login_history").Aggregate(ctx, pipeline)
	if err != nil {
		middleware.LogError(err, "Failed to get engagement metrics")
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		middleware.LogError(err, "Failed to decode engagement metrics")
		return nil, err
	}

	activeLastMonth := 0
	if len(results) > 0 {
		activeLastMonth = int(results[0]["active_last_30_days"].(int32))
	}

	engagementRate := 0.0
	if len(referredUsers) > 0 {
		engagementRate = float64(activeLastMonth) / float64(len(referredUsers)) * 100
	}

	metrics := map[string]interface{}{
		"referral_code":        referralCode,
		"total_referred":       len(referredUsers),
		"active_users":         activeCount,
		"active_last_30_days":  activeLastMonth,
		"engagement_rate":      fmt.Sprintf("%.2f%%", engagementRate),
	}

	return metrics, nil
}

// ==================== PING TEST ====================

func TestPing() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result bson.M
	err := database.DbAuth.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result)

	if err != nil {
		middleware.LogError(err, "MongoDB Ping Failed")
		return err
	}

	fmt.Println(" Pinged your deployment. You successfully connected to MongoDB!")
	return nil
}