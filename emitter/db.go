package emitter

import (
	"context"
	"database/sql"
	"fmt"
	l "pgnextgenconsumer/config"
	_mapper "pgnextgenconsumer/mappers"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

const (
	insertQuery    = "INSERT INTO RequestLog (merchantId, reqType, uri, source, request, response, addedOn) VALUES %s"
	placeholderStr = "(?, ?, ?, ?, ?, ?, ?)"
)

type Database struct {
	Db *sql.DB
}

func InitDB(config l.MySqlConfig) *Database {
	ctx := context.Background()
	connStr := config.Username + ":" + config.Password + "@/" + config.Database
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		l.H.Error(ctx, "Connect with database failed ", err)
		return nil
	}
	db.SetConnMaxLifetime(time.Minute * time.Duration(config.MaxConnLifetime))
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)

	err = db.Ping()
	if err != nil {
		l.H.Error(ctx, "Connect with database failed ", err)
		return nil
	}
	l.H.Info(ctx, "Connect with database established")
	return &Database{Db: db}
}

func (_this Database) Emit(event _mapper.Event) error {
	ctx := context.Background()
	l.H.Info(ctx, "sending data to Emitter failed ")
	stmtIns, err := _this.Db.Prepare(insertQuery)
	if err != nil {
		l.H.Error(ctx, "Failed in preparing query ", err)
		return err
	}
	defer stmtIns.Close()

	result, err := stmtIns.Exec(event.MerchantId, "event.Request", "event.Response", "2022-02-20 21:45")
	if err != nil {
		l.H.Error(ctx, "failed in executing insert query", err)
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		l.H.Error(ctx, "failed in executing ", err)
		return err
	}
	l.H.Info(ctx, "Added new requests to database: ", zap.Int64("rows", n))
	return nil
}

func (_this Database) BatchEmit(events []_mapper.Event) error {
	ctx := context.Background()
	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*7)
	for _, event := range events {
		valueStrings = append(valueStrings, placeholderStr)
		valueArgs = append(valueArgs, event.MerchantId)
		valueArgs = append(valueArgs, event.ReqType)
		valueArgs = append(valueArgs, event.URI)
		valueArgs = append(valueArgs, event.Source)
		valueArgs = append(valueArgs, event.Request)
		valueArgs = append(valueArgs, event.Response)
		valueArgs = append(valueArgs, event.AddedOn)
		fmt.Println(valueArgs...)
	}
	stmt := fmt.Sprintf(insertQuery, strings.Join(valueStrings, ","))
	l.H.Info(ctx, "sending data to database, ", zap.String("query", stmt), zap.Any("values", valueArgs))
	result, err := _this.Db.Exec(stmt, valueArgs...)
	if err != nil {
		l.H.Error(ctx, "Failed in preparing query ", err)
		return err
	}
	l.H.Info(ctx, "Insertion successful, ", zap.Any("result", result))
	return nil
}
