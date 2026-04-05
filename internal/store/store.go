package store

import "htui/internal/types"

// Store — интерфейс хранилища запросов.
// Позволяет подменять реализацию в тестах.
type Store interface {
	// List возвращает все запросы, отсортированные по UpdatedAt desc.
	List() ([]types.SavedRequest, error)

	// Get возвращает запрос по ID или ошибку если не найден.
	Get(id string) (types.SavedRequest, error)

	// Save сохраняет запрос (создаёт или обновляет).
	// Перед сохранением обновляет UpdatedAt.
	Save(r types.SavedRequest) error

	// Delete удаляет запрос по ID.
	Delete(id string) error

	// IsFirstRun возвращает true если приложение запускается впервые
	// (sentinel-файл .seeded не существует).
	IsFirstRun() (bool, error)

	// MarkSeeded создаёт sentinel-файл .seeded, предотвращая повторное сидирование.
	MarkSeeded() error
}
