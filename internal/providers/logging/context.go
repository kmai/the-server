package logging

import "strconv"

type contextKey uint8

func (c contextKey) String() string {
	return "server/context/logging/" + strconv.Itoa(int(c))
}
