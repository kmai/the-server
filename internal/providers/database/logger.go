package database

import (
	"context"
	"errors"
	"time"

	"github.com/kmai/the-server/internal/providers/logging"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type GormLogger struct {
	*logrus.Logger
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	DebugEnabled          bool
}

//nolint:ireturn,nolintlint
func (gl *GormLogger) LogMode(_ logger.LogLevel) logger.Interface {
	return gl
}

func (gl *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	//nolint:asasalint,nolintlint
	gl.Logger.WithContext(ctx).Infof(s, i...)
}

func (gl *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	//nolint:asasalint,nolintlint
	gl.Logger.WithContext(ctx).Warnf(s, i...)
}

func (gl *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	//nolint:asasalint,nolintlint
	gl.Logger.WithContext(ctx).Errorf(s, i...)
}

func (gl *GormLogger) Debug(ctx context.Context, s string, i ...interface{}) {
	//nolint:asasalint,nolintlint
	gl.Logger.WithContext(ctx).Debugf(s, i...)
}

func (gl *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := logrus.Fields{}

	if gl.SourceField != "" {
		fields[gl.SourceField] = utils.FileWithLineNum()
	}

	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && gl.SkipErrRecordNotFound) {
		fields[logrus.ErrorKey] = err
		gl.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)

		return
	}

	if gl.SlowThreshold != 0 && elapsed > gl.SlowThreshold {
		gl.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)

		return
	}

	if gl.DebugEnabled {
		gl.WithContext(ctx).WithFields(fields).Debugf("%s [%s]", sql, elapsed)
	}
}

func NewLogger(ctx context.Context) *GormLogger {
	return &GormLogger{
		DebugEnabled: true,
		Logger:       logging.GetLoggerFromContext(ctx),
	}
}

// Print implements the interface.
func (gl *GormLogger) Print(values ...interface{}) {
	switch values[0] {
	case "sql":
		gl.WithFields(
			logrus.Fields{
				"module":        "gorm",
				"type":          "sql",
				"rows_returned": values[5],
				"src":           values[1],
				"values":        values[4],
				"duration":      values[2],
			},
		).Info(values[3])
	case "log":
		gl.WithFields(logrus.Fields{"module": "gorm", "type": "log"}).Print(values[2])
	}
}
