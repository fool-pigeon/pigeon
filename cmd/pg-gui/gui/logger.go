// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package gui

import (
	"log/slog"
	"os"
	"time"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogFile() error {
	if err := fsx.MkdirAll(Paths.Log()); err != nil {
		return err
	}
	if err := fsx.CreateFileIfNotExists(Paths.LogFile()); err != nil {
		return err
	}
	return nil
}

func NewLogger() *slog.Logger {
	fileWriter := &lumberjack.Logger{
		Filename:   Paths.LogFile(),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
		LocalTime:  true,
	}
	timefmt := time.DateTime + ".000Z07:00"
	return slog.New(
		slog.NewMultiHandler(
			tint.NewHandler(fileWriter, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: timefmt,
				NoColor:    true,
			}),
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: timefmt,
				NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
			}),
		),
	)
}
