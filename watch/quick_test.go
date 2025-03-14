package watch

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/eensymachines.in/telegram-scaffold/models"
	"github.com/eensymachines.in/telegram-scaffold/send"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestWatchUpdates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mybot := &models.MyBot{
		FetchUpdateInterval: 20 * time.Second,
		Id:                  int64(0),
		Token:               "replace_this_with_actual_token",
	}
	updates := WatchUpdates(ctx, mybot, NinegagFwdMsgParser, 30*time.Second, uint16(10))
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata", "localhost", "manjaro-dev", "m4nj4r0-d3v", "botrunjun", "5432")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil || db == nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	send.MassFwdAsReceived(ctx, mybot, updates, send.NewDBFinder(db))
	t.Log("Now closing test")
}
