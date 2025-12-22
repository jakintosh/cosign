package database

type DBKeyStore struct{}

func NewKeyStore() DBKeyStore { return DBKeyStore{} }

func (DBKeyStore) InsertKey(
	id string,
	salt string,
	hash string,
) error {
	_, err := db.Exec(`
		INSERT INTO api_key (id, salt, hash, created)
		VALUES (?1, ?2, ?3, unixepoch());`,
		id,
		salt,
		hash,
	)
	return err
}

func (DBKeyStore) FetchKey(
	id string,
) (
	salt string,
	hash string,
	err error,
) {
	row := db.QueryRow(`
		SELECT salt, hash
		FROM api_key
		WHERE id=?1;`,
		id,
	)

	err = row.Scan(&salt, &hash)
	if err != nil {
		return "", "", err
	}

	_, _ = db.Exec(`
		UPDATE api_key
		SET last_used=unixepoch()
		WHERE id=?1;`,
		id,
	)

	return
}

func (DBKeyStore) DeleteKey(
	id string,
) error {
	_, err := db.Exec(`
		DELETE FROM api_key
		WHERE id=?1;`,
		id,
	)
	return err
}

func (DBKeyStore) CountKeys() (int, error) {
	row := db.QueryRow(`
		SELECT COUNT(*)
		FROM api_key;
	`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
