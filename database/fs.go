package database

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDbFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "block.db")
}

func initDataDirIfNotExists(dataDir string) error {
	if fileExist(getGenesisJsonFilePath(dataDir)) {
		return nil
	}
	dbDir := getDatabaseDirPath(dataDir)
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return err
	}
	gen := getGenesisJsonFilePath(dataDir)
	if err := writeGenesisToDisk(gen); err != nil {
		return err
	}
	blocks := getBlocksDbFilePath(dataDir)
	if err := writeEmptyBlocksDbToDisk(blocks); err != nil {
		return err
	}
	return nil
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func writeEmptyBlocksDbToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(""), os.ModePerm)
}
