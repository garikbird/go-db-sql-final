package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // Подключение к существующей БД
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// get
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)
	require.Equal(t, parcel.Status, storedParcel.Status)
	require.Equal(t, parcel.Address, storedParcel.Address)
	require.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// check deletion
	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// set status
	newStatus := ParcelStatusDelivered
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, updatedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	parcelMap := make(map[int]Parcel)
	for _, parcel := range parcels {
		id, err := store.Add(parcel)
		require.NoError(t, err)
		parcel.Number = id
		parcelMap[id] = parcel
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	for _, storedParcel := range storedParcels {
		// Получаем ожидаемую посылку
		expectedParcel, ok := parcelMap[storedParcel.Number]
		require.True(t, ok)

		// Обнуляем поле Number у обеих структур для сравнения
		storedParcel.Number = 0
		expectedParcel.Number = 0

		// Сравниваем структуры целиком
		require.Equal(t, expectedParcel, storedParcel)
	}
}
