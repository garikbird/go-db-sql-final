package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec(`INSERT INTO parcel (client, status, address, created_at)
	VALUES (?, ?, ?, ?)`, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil

}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	row := s.db.QueryRow(`
		SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`,
		number)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		fmt.Println(err)
		return p, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(`
		SELECT number, client, status, address, created_at
		FROM parcel
		WHERE client = ?`, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Parcel

	for rows.Next() {
		p := Parcel{}

		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(`
		UPDATE parcel
		SET status = ?
		WHERE number = ?`, status, number)
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) GetStatus(number int) (string, error) {
	p := Parcel{}
	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number", sql.Named("number", number))
	err := row.Scan(&p.Status)
	if err != nil {
		return "", fmt.Errorf("select failed: %w", err)
	}

	return p.Status, nil
}
func (s ParcelStore) SetAddress(number int, address string) error {
	status, err := s.GetStatus(number)
	if err != nil {
		return fmt.Errorf("getStatus failed: %w", err)
	}

	if status != ParcelStatusRegistered {
		return fmt.Errorf("address shouldn't be updated")
	}
	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil

}

func (s ParcelStore) Delete(number int) error {
	result, err := s.db.Exec("DELETE FROM parcel WHERE number =:number AND status =:status", sql.Named("number", number), sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	fmt.Println(rowsAffected)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("посылка с номером %d не найдена или ее статус не 'registered'", number)
	}
	return nil
}
