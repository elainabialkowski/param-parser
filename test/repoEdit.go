package main

import (
	"strconv"
	"strings"
)

type RepoEdit struct {
	Id     uint64
	UserId string
}

func (RepoEdit) populate(params string) (RepoEdit, error) {
	var err error
	split := strings.Fields(params)
	populated := RepoEdit{}

	populated.Id, err = strconv.ParseUint(split[0], 10, 64)
	if err != nil {
		return RepoEdit{}, err
	}

	populated.UserId = split[1]

	return populated, nil
}
