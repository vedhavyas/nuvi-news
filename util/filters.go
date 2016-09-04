package util

import (
	"fmt"

	"github.com/keimoon/gore"
)

//FilterMap filters the map contents based on the list taken from redis key
func FilterMap(conn *gore.Conn, key string, fullList map[string]string) (map[string]string, error) {
	reply, err := gore.NewCommand("LRANGE", key, "0", "-1").Run(conn)
	if err != nil {
		return fullList, err
	}

	finishedList := []string{}
	err = reply.Slice(&finishedList)
	if err != nil {
		return fullList, err
	}

	for _, finishedFile := range finishedList {
		_, ok := fullList[finishedFile]
		if ok {
			fmt.Println("skipping", finishedFile)
			delete(fullList, finishedFile)
		}
	}

	return fullList, nil
}
