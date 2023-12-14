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
func VerifyCredentialsAndCreateIfNotExists(partnerName, partnerPassword string, db *sql.DB) error {

	// Search for partner in authenticator table
	var encodedPassword string
	searchPartnerErr := db.QueryRow(util.ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)

	// If partner name is not recorded, add partner with encoded password
	if searchPartnerErr == sql.ErrNoRows {
		// Encode the given password to make a new entry
		encodedPartnerPassword, err := bcrypt.GenerateFromPassword([]byte(partnerPassword), bcrypt.MinCost)
		if err != nil {
			return err
		}
		// Create partner entry into the database
		_, txErr := db.Exec(util.InsertPartnerToAuthSQLCmd, partnerName, encodedPartnerPassword)
		if txErr != nil {
			return txErr
		}
		return nil
	}
	// If partner is found then check if password matches
	err := bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(partnerPassword))
	if err != nil {
		return fmt.Errorf(util.InvalidPasswordErr)
	}
	return nil
}
