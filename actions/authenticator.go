package actions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
)

func authenticatePostRequest(r *http.Request, tx *sql.Tx) (string, error) {
	partnerName := r.FormValue(PartnerNameInputName)
	decodedPassword := r.FormValue(EncodedPasswordInputName)

	// If partner name or password are empty, make partner anonymous
	if partnerName == "" || decodedPassword == "" {
		return "", nil
	}

	// Search for partner in authenticator talbe
	var encodedPassword string
	err := tx.QueryRow(ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)
	// Encode given decoded password
	encodedDecodedPassword := base64.StdEncoding.EncodeToString([]byte(decodedPassword))

	// If partner name is not recorded, add partner with encoded password
	if err == sql.ErrNoRows {
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
	if encodedPassword != encodedDecodedPassword {
		return "", fmt.Errorf(InvalidPasswordErr)
	}
	return partnerName, nil
}

func authenticateGetRequest(r *http.Request, db *sql.DB) (string, error) {
	partnerName := r.FormValue(PartnerNameInputName)
	decodedPassword := r.FormValue(EncodedPasswordInputName)

	// Search for partner in authenticator talbe
	var encodedPassword string
	err := db.QueryRow(ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)
	if err != nil {
		return "", err
	}

	// Encode given decoded password
	encodedDecodedPassword := base64.StdEncoding.EncodeToString([]byte(decodedPassword))
	if encodedPassword != encodedDecodedPassword {
		return "", fmt.Errorf(InvalidPasswordErr)
	}

	return partnerName, nil
}
