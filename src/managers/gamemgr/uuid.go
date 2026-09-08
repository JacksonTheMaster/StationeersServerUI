package gamemgr

import (
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/google/uuid"
)

var GameServerUUID uuid.UUID

func clearGameServerUUID() {
	GameServerUUID = uuid.Nil
	clearGameServerLogFilePath()
}

func createGameServerUUID() {
	GameServerUUID = uuid.New()
	setGameServerLogFilePath(GameServerUUID)
	logger.Core.Debug("Created Game Server with internal UUID: " + GameServerUUID.String())
}
