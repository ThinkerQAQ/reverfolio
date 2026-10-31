package echo

import "fmt"

func (s *Store) migrateSiteAndArchives() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version >= 3 {
		return nil
	}
	if _, err := s.db.Exec(`PRAGMA foreign_keys=OFF`); err != nil {
		return err
	}
	defer s.db.Exec(`PRAGMA foreign_keys=ON`)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`CREATE TABLE registry_next(id INTEGER PRIMARY KEY,title TEXT NOT NULL,repository TEXT COLLATE NOCASE UNIQUE,github_id INTEGER UNIQUE,kind TEXT NOT NULL DEFAULT 'project' CHECK(kind IN ('project','article','site')),canonical_url TEXT UNIQUE);
 INSERT INTO registry_next SELECT id,title,repository,github_id,kind,canonical_url FROM works;
 DROP TABLE works;
 ALTER TABLE registry_next RENAME TO works;
 CREATE TABLE raw_observations(snapshot_id INTEGER NOT NULL REFERENCES snapshots(id),endpoint TEXT NOT NULL,payload TEXT NOT NULL,PRIMARY KEY(snapshot_id,endpoint));
 PRAGMA user_version=3;`)
	if err != nil {
		return err
	}
	rows, err := tx.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	broken := rows.Next()
	rowErr := rows.Err()
	rows.Close()
	if rowErr != nil {
		return rowErr
	}
	if broken {
		return fmt.Errorf("migration would break foreign keys")
	}
	return tx.Commit()
}
