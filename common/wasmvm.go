package common

import ( 
	"errors"
	"liberdade.bsb.br/baas/scripting/database"
)

func RunWasmAction(appId int, userId int, actionScript string, inputData string, connection *database.Conn) (string, error) {
	return "", errors.New("not implemented yet!")
}

