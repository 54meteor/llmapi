package util

import (
	"context"
	"one-api/common/logger"
	"one-api/model"
)

func ReturnPreConsumedQuota(ctx context.Context, preConsumedQuota int, tokenId int) {
	if preConsumedQuota != 0 {
		go func(ctx context.Context) {
			// return pre-consumed quota to database
			err := model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
			if err != nil {
				logger.Error(ctx, "error return pre-consumed quota: "+err.Error())
				return
			}
			// Sync Redis cache with database
			token, err := model.GetTokenById(tokenId)
			if err != nil {
				logger.Error(ctx, "error get token for cache update: "+err.Error())
				return
			}
			err = model.CacheUpdateUserQuota(token.UserId)
			if err != nil {
				logger.Error(ctx, "error update user quota cache: "+err.Error())
			}
		}(ctx)
	}
}
