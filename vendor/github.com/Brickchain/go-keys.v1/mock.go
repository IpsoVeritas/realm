package keys

type mockStoredKeyService struct {
	keys map[string]*StoredKey
}

// NewMockStoredKeyService returns a mock version of StoredKeyService
func NewMockStoredKeyService() StoredKeyService {
	return &mockStoredKeyService{
		keys: make(map[string]*StoredKey),
	}
}

func (m *mockStoredKeyService) Get(id string) (*StoredKey, error) {
	k, ok := m.keys[id]
	if !ok {
		return nil, ErrNoSuchKey
	}

	return k, nil
}

func (m *mockStoredKeyService) Save(k *StoredKey) error {
	m.keys[k.ID] = k

	return nil
}
