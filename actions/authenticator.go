package actions

import (
	"database/sql"
	"os"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func connectToDB() (*sql.DB, error) {
	DBUsername := os.Getenv("DB_USER")
	DBPassword := os.Getenv("DB_PASSWORD")
	DBURL := os.Getenv("DB_URL")
	DBPort := os.Getenv("DB_PORT")

	DBConnStr := DBUsername + ":" + DBPassword + "@tcp(" + DBURL + ":" + DBPort + ")/"
	db, err := sql.Open("mysql", DBConnStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func authenticatePostRequest(r *http.Request, tx *sql.Tx) (string, error) {
	partnerName := r.FormValue(PartnerNameInputName)
	decodedPassword := r.FormValue(DedcodedPasswordInputName)

	// If partner name or password are empty, make partner anonymous
	if partnerName == "" || decodedPassword == "" {
		return "", nil
	}

	// Search for partner in authenticator table
	var encodedPassword string
	searchPartnerErr := tx.QueryRow(ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)
	// Encode given decoded password
	encodedDecodedPassword, err := bcrypt.GenerateFromPassword([]byte(decodedPassword), bcrypt.MinCost)

	// If partner name is not recorded, add partner with encoded password
	if searchPartnerErr == sql.ErrNoRows {
		_, txErr := tx.Exec(InsertPartnerToAuthSQLCmd, partnerName, encodedDecodedPassword)
		if txErr != nil {
			handleTransactionRollback(tx, AuthError, err)
			return "", txErr
		}
		return partnerName, nil
	}
	if err != nil {
		handleTransactionRollback(tx, AuthError, err)
		return "", err
	}

	// If partner is recorded and password is wrong throw data
	err = bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(decodedPassword))
	if err != nil {
		return "", fmt.Errorf(InvalidPasswordErr)
	}
	return partnerName, nil
}

func authenticateGetRequest(r *http.Request, db *sql.DB) (string, error) {
	partnerName := r.FormValue(PartnerNameInputName)
	decodedPassword := r.FormValue(DedcodedPasswordInputName)

	// Search for partner in authenticator table
	var encodedPassword string
	err := db.QueryRow(ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)
	if err != nil {
		return "", err
	}

	// Compare encoded and given passwords
	err = bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(decodedPassword))
	if err != nil {
		return "", fmt.Errorf(InvalidPasswordErr)
	}

	return partnerName, nil
}
