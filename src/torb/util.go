package main

func contains(ids []int64, id int64) bool {
	for k := range ids {
		if ids[k] == id {
			return true
		}
	}
	return false
}
