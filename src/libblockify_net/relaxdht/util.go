package relaxdht

func JsonGetField(json interface{},field string) (interface{},bool){
	m,ok := json.(map[string]interface{})
	if !ok { return nil,false }
	v,ok := m[field]
	return v,ok
}

func JsonGetFieldString(json interface{},field string) (string,bool){
	m,ok := json.(map[string]interface{})
	if !ok { return "",false }
	v,ok := m[field]
	if !ok { return "",false }
	s,ok := v.(string)
	return s,ok
}
