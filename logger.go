/*
 * Author: Anton Volokha
 * Copyright (c) 2020
 */

package main

import (
  "log"
  "os"
  "sync"
  "time"
  "fmt"
)

const (
  INFO = "info.log"
  DANGER = "danger.log"
  ERROR = "error.log"
)

type Logger struct {
  filename string
  *log.Logger
}

var once sync.Once
var info *Logger
var danger *Logger
var err *Logger

// start logger
func setupLogger() {
  once.Do(func() {
    info = createLogger(INFO)
    danger = createLogger(DANGER)
    err = createLogger(ERROR)
  })
}

func createLogger(fname string) *Logger {
  file, _ := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)

  return &Logger{
      filename: fname,
      Logger:   log.New(file, fmt.Sprintf("%s\t", time.Now()), log.Lshortfile),
  }
}

/**
 * Write logs to file
 */
 func infoLog(c *CoinData) {
  str := fmt.Sprintf("usd: %.2f\t eur: %.2f\t uah: %.2f", c.Usd, c.Eur, c.Uah)

  fmt.Println(str)
  info.Println(str)
}

/**
 * Write dander log
 */
func dangerLog(c *CoinData) {
  str := fmt.Sprintf("\nSELL SELL SELL usd: %.2f\t eur: %.2f\t uah: %.2f", c.Usd, c.Eur, c.Uah)

  fmt.Println(str)
  danger.Println(str)
}

/**
 * Write error log
 */

func errorLog(e error) {
  fmt.Println(e)
  err.Println(e)
}
