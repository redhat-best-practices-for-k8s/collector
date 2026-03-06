package api

import (
	"database/sql"
	"fmt"

	"github.com/redhat-best-practices-for-k8s/collector/util"
	"golang.org/x/crypto/bcrypt"
)

func checkPassword(encodedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(plainPassword))
	if err != nil {
		return fmt.Errorf(util.InvalidPasswordErr)
	}
	return nil
}

func CheckIfValidCredentials(partnerName, decodePassword string, db *sql.DB) error {
	var encodedPassword string
	err := db.QueryRow(util.ExtractPartnerAndPasswordCmd, partnerName).Scan(&encodedPassword)
	if err != nil {
		return fmt.Errorf(util.InvalidUsernameErr)
	}
	return checkPassword(encodedPassword, decodePassword)
}

// Already non-empty partner name and decoded password are given
func VerifyCredentialsAndCreateIfNotExists(partnerName, partnerPassword string, db *sql.DB) error {
	// if partner name and password aren't given, post anonymously
	if partnerName == "" && partnerPassword == "" {
		return nil
	}

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
	return checkPassword(encodedPassword, partnerPassword)
}
