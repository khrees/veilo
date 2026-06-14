package repositories_test

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	// Use SQLite for testing - in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Create tables manually without PostgreSQL-specific defaults
	err = db.Exec(`CREATE TABLE "domains" (
		"id" TEXT PRIMARY KEY,
		"name" TEXT NOT NULL UNIQUE,
		"verified" INTEGER NOT NULL DEFAULT 0,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"deleted_at" DATETIME
	)`).Error
	if err != nil {
		panic(err)
	}

	err = db.Exec(`CREATE TABLE "aliases" (
		"id" TEXT PRIMARY KEY,
		"address" TEXT NOT NULL UNIQUE,
		"slug" TEXT NOT NULL,
		"domain" TEXT NOT NULL,
		"real_email" TEXT NOT NULL,
		"display_name" TEXT,
		"label" TEXT,
		"enabled" INTEGER NOT NULL DEFAULT 1,
		"forward_count" INTEGER NOT NULL DEFAULT 0,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"last_used_at" DATETIME,
		"expires_at" DATETIME,
		"max_forwards" INTEGER
	)`).Error
	if err != nil {
		panic(err)
	}

	err = db.Exec(`CREATE TABLE "reply_tokens" (
		"token" TEXT PRIMARY KEY,
		"alias_id" TEXT NOT NULL,
		"original_sender" TEXT NOT NULL,
		"original_subject" TEXT,
		"thread_id" TEXT,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"expires_at" DATETIME NOT NULL
	)`).Error
	if err != nil {
		panic(err)
	}

	err = db.Exec(`CREATE TABLE "forward_logs" (
		"id" TEXT PRIMARY KEY,
		"alias_id" TEXT NOT NULL,
		"direction" TEXT NOT NULL,
		"sender" TEXT,
		"subject" TEXT,
		"status" TEXT NOT NULL,
		"trackers_blocked" INTEGER NOT NULL DEFAULT 0,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error
	if err != nil {
		panic(err)
	}

	return db
}
