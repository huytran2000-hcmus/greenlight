package data

type Permissions []string

func (p Permissions) Include(code string) bool {
	for _, permission := range p {
		if code == permission {
			return true
		}
	}

	return false
}
