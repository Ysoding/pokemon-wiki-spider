package spider

type TempData struct {
	data map[string]interface{}
}

func (t *TempData) Get(key string) interface{} {
	return t.data[key]
}

func (t *TempData) Set(key string, value interface{}) error {
	if t.data == nil {
		t.data = make(map[string]interface{}, 8)
	}
	t.data[key] = value
	return nil
}
