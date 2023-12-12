package api

import (
	"database/sql"
	"fmt"

	"github.com/test-network-function/collector/util"
	"golang.org/x/crypto/bcrypt"
)

func CheckIfValidCredentials(partnerName, decodePassword string, db *sql.DB) error {
	// Search for partner in authenticator table
	var encodedPassword string
	err := db.QueryRow(util.ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)
	if err != nil {
		return fmt.Errorf(util.InvalidUsernameErr)
	}

	// Compare encoded and given passwords
	err = bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(decodePassword))
	if err != nil {
		return fmt.Errorf(util.InvalidPasswordErr)
	}

	return nil
}

// Already non-empty partner name and decoded password are given
func CreateCredentialsIfNotExists(partnerName, decodedPassword string, tx *sql.Tx) error {
	// Search for partner in authenticator table
	var encodedPassword string
	searchPartnerErr := tx.QueryRow(util.ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)
	// Encode given decoded password
	encodedDecodedPassword, err := bcrypt.GenerateFromPassword([]byte(decodedPassword), bcrypt.MinCost)

	// If partner name is not recorded, add partner with encoded password
	if searchPartnerErr == sql.ErrNoRows {
		_, txErr := tx.Exec(util.InsertPartnerToAuthSQLCmd, partnerName, encodedDecodedPassword)
		if txErr != nil {
			util.HandleTransactionRollback(tx, util.AuthError, err)
			return txErr
		}
		return nil
	}
	if err != nil {
		util.HandleTransactionRollback(tx, util.AuthError, err)
		return err
	}

	// If partner is recorded and password is wrong throw data
	err = bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(decodedPassword))
	if err != nil {
		return fmt.Errorf(util.InvalidPasswordErr)
	}
	return nil
}
