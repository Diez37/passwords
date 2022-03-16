package repeater

import (
	"context"
	"github.com/Diez37/passwords/application/blocker"
	"github.com/Diez37/passwords/infrastructure/config"
	"github.com/diez37/go-packages/log"
	"time"
)

func Serve(ctx context.Context, blockerConfig *config.Blocker, logger log.Logger, blocker blocker.Blocker) {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	logger.Info("repeater: started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("repeater: shutdown")
			return
		case <-time.After(blockerConfig.BlockInterval):
			logger.Info("repeater: passwords blocking")

			if err := blocker.Block(ctx); err != nil {
				logger.Error(err)
			}
		}
	}
}
