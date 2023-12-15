package util

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func WriteError(w http.ResponseWriter, context, err string) {
	_, writeErr := w.Write([]byte(err + "\n"))
	if writeErr != nil {
		logrus.Errorf(WritingResponseErr, writeErr)
	}
	logrus.Errorf(context, err)
}

func getClaimFile(w http.ResponseWriter, r *http.Request) multipart.File {
	err := r.ParseMultipartForm(ParseLowerBound << ParseUpperBound)
	if err != nil {
		WriteError(w, RequestContentTypeErr, err.Error())
		return nil
	}

	claimFile, _, err := r.FormFile(ClaimFileInputName)
	if err != nil {
		WriteError(w, FormFileErr, err.Error())
		return nil
	}
	return claimFile
}

func ReadClaimFile(w http.ResponseWriter, r *http.Request) []byte {
	claimFile := getClaimFile(w, r)
	defer claimFile.Close()

	claimFileBytes, err := io.ReadAll(claimFile)
	if err != nil {
		WriteError(w, ReadingFileErr, err.Error())
		return nil
	}

	return claimFileBytes
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetDatabaseEnvVars() (user, password, dbUrl, port string) {
	user = getEnv("DB_USER", "admin")
	password = getEnv("DB_PASSWORD", "collector1")
	dbUrl = getEnv("DB_URL", "collector-db.cn9luyhgvfkp.us-east-1.rds.amazonaws.com")
	port = getEnv("DB_PORT", "3306")

	return user, password, dbUrl, port
}

func GetServerEnvVars() (readTimeOutInt, writeTimeOutInt int, addr, err string) {

	readTimeOutStr := getEnv("SERVER_READ_TIMEOUT", "20")
	writeTimeOutStr := getEnv("SERVER_WRITE_TIMEOUT", "20")
	adrr := getEnv("SERVER_ADDR", ":80")

	// Convert read and write time outs to integers.
	readTimeOutInt, atoiErr := strconv.Atoi(readTimeOutStr)
	if atoiErr != nil {
		return -1, -1, "", atoiErr.Error()
	}

	writeTimeOutInt, atoiErr = strconv.Atoi(writeTimeOutStr)
	if atoiErr != nil {
		return -1, -1, "", atoiErr.Error()
	}

	return readTimeOutInt, writeTimeOutInt, adrr, ""
}

func GetS3ConnectEnvVars() (string, string) {
	return getEnv("AWS_ACCESS_KEY", ""), getEnv("AWS_SECRET_ACCESS_KEY", "")
}
