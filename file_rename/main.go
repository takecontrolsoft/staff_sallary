package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/takecontrolsoft/go_multi_log/logger"
	"github.com/takecontrolsoft/go_multi_log/logger/levels"
	"github.com/takecontrolsoft/go_multi_log/logger/loggers"
	"github.com/xuri/excelize/v2"
)

func main() {
	fl, shouldReturn := RegisterLogger()
	if shouldReturn {
		return
	}
	defer func() {
		fl.Stop()
	}()

	m, shouldReturn := ReadExcel("Clients.xlsx", "Clients", "ЕИК", "ИМЕ")
	if shouldReturn {
		return
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Getting home path failed")
		return
	}
	microDir := filepath.Join(dirname, "Documents", "Microinvest")
	subitems, err := os.ReadDir(microDir)
	if err != nil {
		logger.Error("Getting files from current folder failed")
		return
	}
	for _, subitem := range subitems {
		fn := subitem.Name()
		fileExtension := filepath.Ext(fn)
		if !subitem.IsDir() && fileExtension == ".txt" {
			logger.InfoF("File: %s", fn)
			for eik, name := range m {
				if strings.Contains(fn, eik) && !strings.Contains(fn, fmt.Sprintf("%s_%s", name, eik)) {
					newName := strings.ReplaceAll(fn, eik, fmt.Sprintf("%s_%s", name, eik))
					os.Rename(filepath.Join(microDir, fn), filepath.Join(microDir, newName))
					logger.InfoF("File '%s' renamed to '%s' in folder %s", fn, newName, microDir)
				}

			}
		}
	}
	logger.InfoF("File renaming completed successfully")

}

func ReadExcel(excelName string, sheetName string, eik string, name string) (map[string]string, bool) {
	var m = make(map[string]string)
	fe, err := excelize.OpenFile(excelName)
	if err != nil {
		logger.ErrorF("Open file %s failed, Error: %v", excelName, err)
		return nil, true
	}
	defer func() {

		if err := fe.Close(); err != nil {
			logger.ErrorF("Closing file %s failed, Error: %v", excelName, err)
		}
	}()

	rows, err := fe.GetRows(sheetName)
	if err != nil {
		logger.ErrorF("Getting rows from %s failed, Error: %v", sheetName, err)
		return nil, true
	}
	var eikIndex, nameIndex int = 0, 0

	for c, colCell := range rows[0] {
		logger.InfoF("Row headers: %v", colCell)
		if colCell == eik {
			eikIndex = c
		} else if colCell == name {
			nameIndex = c
		}
	}
	logger.InfoF("EIK index: %v, NAME index: %v", eikIndex, nameIndex)

	for r, row := range rows {
		if r > 0 {
			logger.InfoF("Row index: %v", r)
			eikValue := row[eikIndex]
			nameValue := row[nameIndex]

			m[eikValue] = nameValue
			logger.InfoF("EIK: %s, NAME: %s", eikValue, nameValue)
		}
	}
	return m, false
}

func RegisterLogger() (*loggers.FileLogger, bool) {
	fileOptions := loggers.FileOptions{
		Directory:     "./logs",
		FilePrefix:    "file_rename",
		FileExtension: ".log",
	}
	err := os.MkdirAll(fileOptions.Directory, os.ModePerm)
	if err != nil {
		return nil, false
	}
	fl := loggers.NewFileLogger(levels.All, "", fileOptions)

	err = logger.RegisterLogger("file_rename", fl)
	if err != nil {
		logger.ErrorF("Creating log file failed %v", err)
		return nil, true
	}
	return fl, false
}
